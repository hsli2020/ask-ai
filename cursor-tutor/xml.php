<?php
/*
<Packages>
    <ApplicationName>MSYS2 64bit</ApplicationName>
    <ApplicationVersion>1.0.0</ApplicationVersion>
    <Package>
        <Name>com.msys2.root</Name>
        <Title>MSYS2 64bit</Title>
        <Description>Cygwin-derived Posix-like env. for Windows with Arch Linux's Pacman</Description>
        <Version>20210725</Version>
        <LastUpdateDate></LastUpdateDate>
        <InstallDate>2021-10-25</InstallDate>
        <Size>0</Size>
        <Checkable>true</Checkable>
    </Package>
    <Package>
        <Name>com.msys2.root.base</Name>
        <Title>MSYS2 64bit base</Title>
        <Description>Cygwin-derived Posix-like env. for Windows with Arch Linux's Pacman</Description>
        <Version>20210725</Version>
        <LastUpdateDate></LastUpdateDate>
        <InstallDate>2021-10-25</InstallDate>
        <Size>344676172</Size>
        <Checkable>true</Checkable>
    </Package>
</Packages>
*/

$xml = simplexml_load_file("file.xml");
$file = fopen('file.csv', 'w');
$header = array('Name', 'Title', 'Description', 'Version', 'InstallDate', 'Size');
fputcsv($file, $header);
foreach($xml->Package as $package) {
    $row = array(
        (string)$package->Name,
        (string)$package->Title,
        (string)$package->Description,
        (string)$package->Version,
        (string)$package->InstallDate,
        (string)$package->Size
    );
    fputcsv($file, $row);
}
fclose($file);
