<?php
header("Access-Control-Allow-Origin: *");

/**
 * Deps
 */

use \Psr\Http\Message\ServerRequestInterface as Request;
use \Psr\Http\Message\ResponseInterface as Response;

require './vendor/autoload.php';

/**
 * HTTP Handlers
 */

$app = new \Slim\App;

$app->post('/api/v1/upload-session', function (Request $request, Response $response, array $args) {
    $data = [ "foo" => "bar" ];

    return $response->withJson($data);
});

$app->run();

//config definition
$oy_iri = array(
    "IP_GOES_HERE",
    "IP_GOES_HERE",
    "IP_GOES_HERE"
);
$oy_fetch_timeout = 5;

//config definition
function oy_error($oy_error) {
    echo "{\"exception\":\"Invalid request: ".$oy_error."\",\"duration\":0}";
    exit;
}
function oy_iri_select($oy_iri_array) {
    if (count($oy_iri_array)==1) return "http://".$oy_iri_array[0];
    return "http://".$oy_iri_array[array_rand($oy_iri_array)];
}
if (!isset($_POST)) oy_error(1);
$oy_post_raw = file_get_contents("php://input");
$oy_command_array = json_decode($oy_post_raw);
if (!isset($oy_command_array->command)) oy_error(2);
$oy_post = false;//to stop my ide from complaining
if ($oy_command_array->command=="getTransactionsToApprove") {
    if (!is_numeric($oy_command_array->depth)||$oy_command_array->depth<1||$oy_command_array->depth>8) oy_error(3);
    $oy_post = "{\"command\":\"getTransactionsToApprove\", \"depth\":".$oy_command_array->depth."}";
}
else oy_error(4);
$ch = curl_init();
curl_setopt($ch, CURLOPT_URL, oy_iri_select($oy_iri));
curl_setopt($ch, CURLOPT_POST, 1);
curl_setopt($ch, CURLOPT_POSTFIELDS, $oy_post);
curl_setopt($ch, CURLOPT_RETURNTRANSFER, 1);
curl_setopt($ch, CURLOPT_CONNECTTIMEOUT, $oy_fetch_timeout);
$data = curl_exec($ch);
curl_close($ch);
echo $data;
