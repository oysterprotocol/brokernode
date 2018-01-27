<?php

namespace App\Http\Controllers;

use App\Clients\BrokerNode;
use App\DataMap;
use App\UploadSession;
use Illuminate\Http\Request;
use Tuupola\Trytes;

class UploadSessionController extends Controller
{
    /**
     * Display a listing of the resource.
     *
     * @return \Illuminate\Http\Response
     */
    public function index()
    {
        //
    }

    /**
     * Show the form for creating a new resource.
     *
     * @return \Illuminate\Http\Response
     */
    public function create()
    {
        //
    }

    /**
     * Store a newly created resource in storage.
     *
     * @param  \Illuminate\Http\Request  $request
     * @return \Illuminate\Http\Response
     */
    public function store(Request $request)
    {
        $genesis_hash = $request->input('genesis_hash');
        $file_size_bytes = $request->input('file_size_bytes');

        // TODO: Make this an env variable.
        $file_chunk_count = ceil($file_size_bytes / 1000);
        // This could take a while, but if we make this async, we have a race
        // condition if the client attempts to upload before broker-node
        // can save to DB.
        DataMap::buildMap($genesis_hash, $file_chunk_count);

        $upload_session = UploadSession::firstOrCreate([
            'genesis_hash' => $genesis_hash,
            'file_size_bytes' => $file_size_bytes
        ]);

        return response()->json($upload_session);
    }

    /**
     * Display the specified resource.
     *
     * @param  int  $id
     * @return \Illuminate\Http\Response
     */
    public function show($id)
    {
        //
    }

    /**
     * Show the form for editing the specified resource.
     *
     * @param  int  $id
     * @return \Illuminate\Http\Response
     */
    public function edit($id)
    {
        //
    }

    /**
     * Update the specified resource in storage.
     *
     * @param  \Illuminate\Http\Request  $request
     * @param  int  $id
     * @return \Illuminate\Http\Response
     */
    public function update(Request $request, $id)
    {
        $session = UploadSession::find($id);
        if (empty($session)) return response('Session not found.', 404);

        $genesis_hash = $session['genesis_hash'];
        $chunk = $request->input('chunk');

        $data_map = DataMap::where('genesis_hash', $genesis_hash)
            ->where('chunk_idx', $chunk['idx'])
            ->first();

        // Error Responses
        if (empty($data_map)) return response('Datamap not found', 404);

        $trytes = new Trytes(["characters" => Trytes::IOTA]);

        $message_in_tryte_format = $trytes->encode($chunk["data"]);

        $hash_in_tryte_format = $trytes->encode($data_map["hash"]);
        $shortened_hash = substr($hash_in_tryte_format, 0, 81);

        // Process chunk.
        $brokerReq = (object)[
            "responseAddress" =>
                "{$_SERVER['REMOTE_ADDR']}:{$_SERVER['REMOTE_PORT']}",
            "message" => $message_in_tryte_format,
            "chunkId" => $chunk["idx"],
            "address" => $shortened_hash,
        ];


        /*TODO
         * may need to return more stuff from processNewChunk
         */

        $updatedChunk = BrokerNode::processNewChunk($brokerReq);

        // Updates datamap with hooknode url, status, and chunk.
        $data_map->hooknode_id = $updatedChunk->hooknodeUrl;
        $data_map->trunkTransaction = $updatedChunk->trunkTransaction;
        $data_map->branchTransaction = $updatedChunk->branchTransaction;
        $data_map->address = $shortened_hash;
        $data_map->message = $message_in_tryte_format;
        $data_map->chunk = $chunk["data"];
        $data_map->status = DataMap::status['pending'];
        $data_map->save();

        return response('Success.', 204);
    }

    /**
     * Remove the specified resource from storage.
     *
     * @param  int  $id
     * @return \Illuminate\Http\Response
     */
    public function destroy($id)
    {
        // TODO: Delete session & datamap
    }

    /**
     * Gets the status of a chunk. This will be polled until status  is complete
     * or error.
     *
     * @param  int  $body => { genesis_hash, chunk: { idx, hash } }
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
        if (empty($data_map))  return response('Datamap not found', 404);


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

    public function brokerNodeListener(Request $request)
    {
        $cmd = $request->input('command');
        $resAddress = "{$_SERVER['REMOTE_ADDR']}:{$_SERVER['REMOTE_PORT']}";
        // This is a hack to cast an associative array to an object.
        // I don't know how to use PHP properly :(
        $req = (object)[
            "command" => $cmd,
            "responseAddress" => $resAddress,
            "message" => $request->input('message'),
            "chunkId" => $request->input('chunkId'),
            "address" => "WHLOOOOOOAAAAAAAAAAAAAAAAALAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
        ];

        try {
            switch($cmd) {
                case 'processNewChunk':
                    BrokerNode::processNewChunk($request);
                    return response('Success.', 204);

                case 'verifyTx':
                    $dataIsAttached = !BrokerNode::dataNeedsAttaching($request);
                    // TODO: Figure out what client expects.
                    return response()->json(["dataIsAttached" => $dataIsAttached]);

                default:
                    return response("Unrecognized command: {$cmd}", 404);
            }
        } catch (Exception $e) {
            return response("Internal Server Error: {$e->getMessage()}", 500);
        }

    }
}
