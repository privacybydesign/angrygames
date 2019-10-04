<?php
require_once 'config.php';

date_default_timezone_set('UTC');
$protocol = explode(':', IRMA_SERVER_URL, 2)[0];

$sprequests = [
    '18plus' => [
        '@context' => 'https://irma.app/ld/request/disclosure/v2',
        'disclose' => [
            [
                ['pbdf.pbdf.ageLimits.over18'],
                ['pbdf.nijmegen.ageLimits.over18'],
                ['pbdf.gemeente.personalData.over18'],
            ],
        ],
    ],
    '16plus' => [
        '@context' => 'https://irma.app/ld/request/disclosure/v2',
        'disclose' => [
            [
                ['pbdf.pbdf.ageLimits.over16'],
                ['pbdf.nijmegen.ageLimits.over16'],
                ['pbdf.gemeente.personalData.over16'],
            ],
        ],
    ],
    '12plus' => [
        '@context' => 'https://irma.app/ld/request/disclosure/v2',
        'disclose' => [
            [
                ['pbdf.pbdf.ageLimits.over12'],
                ['pbdf.nijmegen.ageLimits.over12'],
                ['pbdf.gemeente.personalData.over12'],
            ],
        ],
    ],
];

function start_session($type) {
    global $sprequests, $protocol;

    if (array_key_exists($type, $sprequests))
        $sessionrequest = $sprequests[$type];
    else
        stop();

    $jsonsr = json_encode($sessionrequest);

    $api_call = array(
        $protocol => array(
            'method' => 'POST',
            'header' => "Content-type: application/json\r\n"
                . "Content-Length: " . strlen($jsonsr) . "\r\n"
                . "Authorization: " . API_TOKEN . "\r\n",
            'content' => $jsonsr
        )
    );

    $resp = file_get_contents(IRMA_SERVER_URL . '/session', false, stream_context_create($api_call));
    if (! $resp) {
        error();
    }
    return $resp;
}

function error() {
    http_response_code(500);
    echo 'Internal server error';
    exit();
}

function stop() {
    http_response_code(400);
    echo 'Invalid request';
    exit();
}

if (!isset($_GET['type']))
    stop();

echo start_session($_GET['type']);
