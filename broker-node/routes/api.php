<?php

use Illuminate\Http\Request;

/*
|--------------------------------------------------------------------------
| API Routes
|--------------------------------------------------------------------------
|
| Here is where you can register API routes for your application. These
| routes are loaded by the RouteServiceProvider within a group which
| is assigned the "api" middleware group. Enjoy building your API!
|
*/

Route::middleware('auth:api')->get('/user', function (Request $request) {
    return $request->user();
});

Route::group(['prefix' => 'v1'], function() {
    Route::resources([
        '/upload-sessions' => 'UploadSessionController',
        '/hooknodes' => 'HookNodeController'
    ]);
    Route::get('/chunk-status', 'UploadSessionController@chunkStatus');

    // This is a temporary route used to integrate BrokerNode into
    // laravel app. This will do exactly what BrokerListener did before.
    Route::post('/broker-node-listener', 'UploadSessionController@brokerNodeListener');
});
