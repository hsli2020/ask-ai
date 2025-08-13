<?php

/*
<?xml version="1.0" encoding="UTF-8"?>
<Network>
    <ProxyType>1</ProxyType>
    <Ftp>
        <Host></Host>
        <Port>0</Port>
        <Username></Username>
        <Password></Password>
    </Ftp>
    <Http>
        <Host></Host>
        <Port>0</Port>
        <Username></Username>
        <Password></Password>
    </Http>
    <Repositories/>
</Network>
*/

$network = array(
    'ProxyType' => 1,
    'Ftp' => array(
        'Host' => '',
        'Port' => 0,
        'Username' => '',
        'Password' => ''
    ),
    'Http' => array(
        'Host' => '',
        'Port' => 0,
        'Username' => '',
        'Password' => ''
    ),
    'Repositories' => array()
);

$xml = new SimpleXMLElement('<?xml version="1.0" encoding="UTF-8"?><Network></Network>');

foreach ($network as $key => $value) {
    if (is_array($value)) {
        $child = $xml->addChild($key);
        foreach ($value as $k => $v) {
            $child->addChild($k, $v);
        }
    } else {
        $xml->addChild($key, $value);
    }
}

echo $xml->asXML();
