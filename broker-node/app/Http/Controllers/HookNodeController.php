<?php

namespace App\Http\Controllers;

use App\HookNode;
use Illuminate\Http\Request;

class HookNodeController extends Controller
{
    /**
     * Store a newly created resource in storage.
     *
     * @param  \Illuminate\Http\Request  $request
     * @return \Illuminate\Http\Response
     */
    public function store(Request $request) {
        $ip_address = $request->input('ip_address');
        HookNode::insertNode($ip_address);

        return response('Success.', 204);
    }
}
