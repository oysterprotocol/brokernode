<?php

namespace App\Http\Controllers;

use Segment;
use \Exception;
use App\Clients\BrokerNode;
use App\DataMap;
use App\UploadSession;
use GuzzleHttp\Client;
use Illuminate\Http\Request;
use Illuminate\Support\Collection;
use Tuupola\Trytes;

class UploadSessionController extends Controller
{
    private static $SegmentStarted = null;

    /**
     * Store a newly created resource in storage.
     *
     * @param  \Illuminate\Http\Request $request
     * @return \Illuminate\Http\Response
     */
    public function store(Request $request)
    {
        $genesis_hash = $request->input('genesis_hash');
        $file_size_bytes = $request->input('file_size_bytes');
        $beta_brokernode_ip = $request->input('beta_brokernode_ip');

        // TODO: Handle PRL Payments.

        // Starts session with beta.
        try {
            $beta_session =
                self::startSessionBeta($genesis_hash, $file_size_bytes, $beta_brokernode_ip);
        } catch (Exception $e) {
            return response("Error: Beta start session failed: {$e}", 500);
        }

        // Starts session on self (alpha).
        $upload_session =
            self::startSession($genesis_hash, $file_size_bytes);

        // Appends beta_session_id for client.
        $res = clone $upload_session;
        $res['beta_session_id'] = $beta_session["id"];

        return response()->json($res);
    }

    private static function startSessionBeta($genesis_hash, $file_size_bytes, $beta_brokernode_ip)
    {
        $beta_broker_path = "{$beta_brokernode_ip}/api/v1/upload-sessions/beta";
        if (!filter_var($beta_broker_path, FILTER_VALIDATE_URL)) {
            return response("Error: Invalid Beta IP {$beta_brokernode_ip}", 422);
        }

        // Starts session with beta.
        $http_client = new Client();
        $res = $http_client->post($beta_broker_path, [
            'form_params' => [
                'genesis_hash' => $genesis_hash,
                'file_size_bytes' => $file_size_bytes
            ]
        ]);
        $beta_session = json_decode($res->getBody(), true);
        return $beta_session;
    }

    /**
     * Store a newly created resource in storage. This is for the beta session
     *
     * @param  \Illuminate\Http\Request $request
     * @return \Illuminate\Http\Response
     */
    public function storeBeta(Request $request)
    {
        $genesis_hash = $request->input('genesis_hash');
        $file_size_bytes = $request->input('file_size_bytes');

        $upload_session =
            self::startSession($genesis_hash, $file_size_bytes, "beta");

        return response()->json($upload_session);
    }

    /**
     * Update the specified resource in storage.
     *
     * @param  \Illuminate\Http\Request $request
     * @param  int $id
     * @return \Illuminate\Http\Response
     */
    public function update(Request $request, $id)
    {
        self::initSegment();

        $session = UploadSession::find($id);
        if (empty($session)) return response('Session not found.', 404);

        $genesis_hash = $session['genesis_hash'];
        $chunks = $request->input('chunks');

        $res_addr = "{$_SERVER['REMOTE_ADDR']}:{$_SERVER['REMOTE_PORT']}";

        // Collect hashes
        // $chunk_idxs = array_map(function ($c) { return $c["idx"]; }, $chunks);
        // $data_maps = DataMap::where('genesis_hash', $genesis_hash)
        //     ->whereIn('chunk_idx', $chunk_idxs)
        //     ->select('chunk_idx', 'hash');
        // unset($chunk_idxs); // Free memory.
        // $idx_to_hash =
        //     $data_maps->reduce(function ($acc, $dmap) {
        //         return $acc[$dmap['idx']] = $dmap['hash'];
        //     }, []);
        // unset($data_maps); // Free memory.

        // Adapt chunks to reqs that hooknode expects.
        $chunk_reqs = collect($chunks)
            ->map(function ($chunk) use ($genesis_hash, $res_addr) {
                // TODO: N queries, optimize later.
                $data_map = DataMap::where('genesis_hash', $genesis_hash)
                    ->where('chunk_idx', $chunk['idx'])
                    ->select('hash')
                    ->first();

                Segment::track([
                    "userId" => "Oyster",
                    "event" => "chunk_sent_from_client",
                    "properties" => [
                        "client_address" => $_SERVER['REMOTE_ADDR'],
                        "chunk_idx" => $chunk['idx']
                    ]
                ]);

                return (object)[
                    'responseAddress' => $res_addr,
                    'address' => self::hashToAddrTrytes($data_map["hash"]),
                    'message' => $chunk['data'],
                    'chunk_idx' => $chunk['idx'],
                ];
            });

        // Process chunks in 1000 chunk batches.
        $chunk_reqs
            ->chunk(1000)// Limited by IRI API.
            ->each(function ($req_list, $idx) {
                $chunks_array = $req_list->all();
                BrokerNode::processChunks($chunks_array);
            });

        // Save to DB.
        $chunk_reqs
            ->each(function ($req) use ($genesis_hash) {
                DataMap::where('genesis_hash', $genesis_hash)
                    ->where('chunk_idx', $req->chunk_idx)
                    ->update([
                        'address' => $req->address,
                        'message' => $req->message,
                    ]);
            });

        return response('Success.', 204);
    }

    /**
     * Remove the specified resource from storage.
     *
     * @param  int $id
     * @return \Illuminate\Http\Response
     */
    public function destroy($id)
    {
        $session = UploadSession::find($id);
        if (empty($session)) return response('Session not found.', 404);

        // TODO: More auth checking? Maybe also send the genesis_hash?

        DataMap::where('genesis_hash', $session['genesis_hash'])->delete();
        $session->delete();

        return response('Success.', 204);
    }

    /**
     * Gets the status of a chunk. This will be polled until status  is complete
     * or error.
     *
     * @param  int $body => { genesis_hash, chunk: { idx, hash } }
     * @return \Illuminate\Http\Response
     */
    public function chunkStatus(Request $request)
    {
        // TODO: Make it middleware to find datamap and check hashes match.
        // It is shared by a few functions.

        $genesis_hash = $request->input('genesis_hash');
        $chunk_idx = $request->input('chunk_idx');

        $data_map = DataMap::where('genesis_hash', $genesis_hash)
            ->where('chunk_idx', $chunk_idx)
            ->first();

        // Error Responses
        if (empty($data_map)) return response('Datamap not found', 404);


        // Don't need to check tangle if already detected to be complete.
        if ($data_map['status'] == DataMap::status['complete']) {
            return response()->json(['status' => $data_map['status']]);
        }

        // Check tangle. This should be done in the background.
        $isAttached = !BrokerNode::dataNeedsAttaching($request);
        if ($isAttached) {
            // Saving to DB is not needed yet, but will be once we check
            // status on the tangle in the background.
            $data_map['status'] = DataMap::status['complete'];
            $data_map->save();
        }

        return response()->json(['status' => $data_map['status']]);
    }

    /**
     * Private
     */

    private static function startSession(
        $genesis_hash, $file_size_bytes, $type = "alpha"
    )
    {
        // TODO: Make 2187 an env variable.
        $file_chunk_count = ceil($file_size_bytes / 2187);
        // This could take a while, but if we make this async, we have a race
        // condition if the client attempts to upload before broker-node
        // can save to DB.
        DataMap::buildMap($genesis_hash, $file_chunk_count);

        return UploadSession::firstOrCreate([
            'type' => $type,
            'genesis_hash' => $genesis_hash,
            'file_size_bytes' => $file_size_bytes
        ]);

        return response()->json($upload_session);
    }

    private static function hashToAddrTrytes($hash)
    {
        $trytes = new Trytes(["characters" => Trytes::IOTA]);
        $hash_in_trytes = $trytes->encode($hash);
        return substr($hash_in_trytes, 0, 81);
    }

    private static function initSegment()
    {
        if (is_null(self::$SegmentStarted)) {
            Segment::init("SrQ0wxvc7jp2XDjZiEJTrkLAo4FC2XdD");
            self::$SegmentStarted = true;
        }
    }
}
