<?php

namespace App\Http\Controllers;

use App\Clients\BrokerNode;
use App\DataMap;
use App\UploadSession;
use GuzzleHttp\Client;
use Illuminate\Http\Request;
use Illuminate\Support\Collection;
use Tuupola\Trytes;

class UploadSessionControllerV2 extends Controller
{
    public function update(Request $request, $id)
    {
        $session = UploadSession::find($id);
        if (empty($session)) return response('Session not found.', 404);

        $genesis_hash = $session['genesis_hash'];
        $chunks = $request->input('chunks');

        $res_addr = "{$_SERVER['REMOTE_ADDR']}:{$_SERVER['REMOTE_PORT']}";

        // Collect hashes
        $chunk_idxs = array_map(function ($c) {
            return $c["idx"];
        }, $chunks);
        $data_maps = DataMap::where('genesis_hash', $genesis_hash)
            ->whereIn('chunk_idx', $chunk_idxs)
            ->select('chunk_idx', 'hash')
            ->get();
        unset($chunk_idxs); // Free memory.
        $idx_to_hash =
            $data_maps->reduce(function ($acc, $dmap) {
                return $acc[$dmap['idx']] = $dmap['hash'];
            }, []);
        unset($data_maps); // Free memory.

        // Adapt chunks to reqs that hooknode expects.
        $chunk_reqs = collect($chunks)
            ->map(function ($chunk, $idx) use ($genesis_hash, $res_addr, $idx_to_hash) {
                $hash = $idx_to_hash[$chunk['idx']];

                return (object)[
                    'responseAddress' => $res_addr,
                    'address' => self::hashToAddrTrytes($hash),
                    'message' => $chunk['data'],
                    'chunkId' => $chunk['idx'],
                ];
            });

        // Process chunks in 1000 chunk batches.
        $chunk_reqs
            ->chunk(1000)// Limited by IRI API.
            ->each(function ($req_list, $idx) {
                BrokerNode::processChunks($req_list);
            });

        // Save to DB.
        $chunk_reqs
            ->each(function ($req, $idx) use ($genesis_hash) {
                DataMap::where('genesis_hash', $genesis_hash)
                    ->where('chunk_idx', $req->chunkId)
                    ->update([
                        'address' => $req->address,
                        'message' => $req->message,
                    ]);
            });

        return response('Success.', 204);
    }

    private static function hashToAddrTrytes($hash)
    {
        $trytes = new Trytes(["characters" => Trytes::IOTA]);
        $hash_in_trytes = $trytes->encode($hash);
        return substr($hash_in_trytes, 0, 81);
    }
}
