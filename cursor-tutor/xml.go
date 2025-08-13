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

type Package struct {
    Name           string `xml:"Name"`
    Title          string `xml:"Title"`
    Description    string `xml:"Description"`
    Version        string `xml:"Version"`
    LastUpdateDate string `xml:"LastUpdateDate"`
    InstallDate    string `xml:"InstallDate"`
    Size           string `xml:"Size"`
    Checkable      string `xml:"Checkable"`
}

type Packages struct {
    ApplicationName    string `xml:"ApplicationName"`
    ApplicationVersion string `xml:"ApplicationVersion"`
    Package            []Package `xml:"Package"`
}

func generateXML() ([]byte, error) {
    
    package1 := Package{}
    
    package1.Name = "com.msys2.root"
    package1.Title = "MSYS2 64bit"
    package1.Description = "Cygwin-derived Posix-like env. for Windows with Arch Linux's Pacman"
    package1.Version = "20210725"
    package1.LastUpdateDate = ""
    package1.InstallDate = "2021-10-25"
    package1.Size = "0"
    package1.Checkable = "true"

    package2 := Package{}
    package2.Name = "com.msys2.root.base"
    package2.Title = "MSYS2 64bit base"
    package2.Description = "Cygwin-derived Posix-like env. for Windows with Arch Linux's Pacman"
    package2.Version = "20210725"
    package2.LastUpdateDate = ""
    package2.InstallDate = "2021-10-25"
    package2.Size = "344676172"
    package2.Checkable = "true"

    packages := Packages{}
    packages.ApplicationName = "MSYS2 64bit"
    packages.ApplicationVersion = "1.0.0"
    packages.Package = []Package{
        package1,
        package2,
    }
    
    return xml.MarshalIndent(packages, "", "    ")
}
