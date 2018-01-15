<?php

namespace App\Http\Controllers;

use App\Clients\HookNode;
use App\DataMapping;
use App\UploadSession;
use Illuminate\Http\Request;

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

        // TODO: Where does $file_chunk_count this come from.
        $file_chunk_count = 3;
        // This could take a while, but if we make this async, we have a race
        // condition if the client attempts to upload before broker-node
        // can save to DB.
        DataMapping::buildMap($genesis_hash, $file_chunk_count);

        $upload_session = UploadSession::create([
            'genesis_hash' => $genesis_hash,
            'file_size_bytes' => $file_size_bytes,
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

        $data_mapping = DataMapping::where('genesis_hash', $genesis_hash)
            ->where('chunk_idx', $chunk['idx'])
            ->first();
        if (empty($data_mapping)) return response('Datamap not found', 404);

        // TODO: What to do with $data_mapping['hash']
        // TODO: Send off $chunk['data'] somewhere to be stored on tangle
        HookNode::processNewChunk($chunk['hash'], $chunk['data'], $chunk['idx'])
            ->then(
                // TODO: What do do with these response?
                function($res) {
                    echo $res;
                },
                function($err) {
                    echo $err;
                }
            );

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
}
