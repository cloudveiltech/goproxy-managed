Function FirstObject($Array) {
    if($Array -is [array] -or $Array.GetType().IsArray) {
        if($Array.Count -gt 0) {
            return $Array[0]
        } else {
            return $null
        }
    } else {
        return $Array
    }
}

Function Find-MsBuild([int] $MaxVersion = 2017)
{
    $agentPath = "$Env:programfiles\Microsoft Visual Studio\2022\Community\MSBuild\Current\Bin\msbuild.exe"
    $devPath = "$Env:programfiles (x86)\Microsoft Visual Studio\2017\Enterprise\MSBuild\15.0\Bin\msbuild.exe"
    $proPath = "$Env:programfiles (x86)\Microsoft Visual Studio\2017\Professional\MSBuild\15.0\Bin\msbuild.exe"
    $communityPath = "$Env:programfiles (x86)\Microsoft Visual Studio\2017\Community\MSBuild\15.0\Bin\msbuild.exe"
    $fallback2015Path = "${Env:ProgramFiles(x86)}\MSBuild\14.0\Bin\MSBuild.exe"
    $fallback2013Path = "${Env:ProgramFiles(x86)}\MSBuild\12.0\Bin\MSBuild.exe"
    $fallbackPath = "C:\Windows\Microsoft.NET\Framework\v4.0.30319"
        
    If ((2017 -le $MaxVersion) -And (Test-Path $agentPath)) { return $agentPath } 
    If ((2017 -le $MaxVersion) -And (Test-Path $devPath)) { return $devPath } 
    If ((2017 -le $MaxVersion) -And (Test-Path $proPath)) { return $proPath } 
    If ((2017 -le $MaxVersion) -And (Test-Path $communityPath)) { return $communityPath } 
    If ((2015 -le $MaxVersion) -And (Test-Path $fallback2015Path)) { return $fallback2015Path } 
    If ((2013 -le $MaxVersion) -And (Test-Path $fallback2013Path)) { return $fallback2013Path } 
    If (Test-Path $fallbackPath) { return $fallbackPath } 

    $pathMsbuild = (which msbuild)
    if ($LastExitCode -ne 0) {
        throw "Yikes - Unable to find msbuild"
    } else {
        return $pathMsbuild
    }
}

function Get-XmlNode([ xml ]$XmlDocument, [string]$NodePath, [string]$NamespaceURI = "", [string]$NodeSeparatorCharacter = '.')
{
    # If a Namespace URI was not given, use the Xml document's default namespace.
    if ([string]::IsNullOrEmpty($NamespaceURI)) { $NamespaceURI = $XmlDocument.DocumentElement.NamespaceURI }   
     
    # In order for SelectSingleNode() to actually work, we need to use the fully qualified node path along with an Xml Namespace Manager, so set them up.
    $xmlNsManager = New-Object System.Xml.XmlNamespaceManager($XmlDocument.NameTable)
    $xmlNsManager.AddNamespace("ns", $NamespaceURI)
    $fullyQualifiedNodePath = "/ns:$($NodePath.Replace($($NodeSeparatorCharacter), '/ns:'))"
     
    # Try and get the node, then return it. Returns $null if the node was not found.
    $node = $XmlDocument.SelectSingleNode($fullyQualifiedNodePath, $xmlNsManager)
    return $node
}

if(!(Test-Path variable:global:IsWindows)) {
    if($Env:OS -eq "Windows_NT") {
        $IsWindows = $true
        $IsMacOS = $false
        $IsLinux = $false
    } else {
        $IsWindows = $false
        $IsMacOS = $false
        $IsLinux = $false
    }
}

$currentLocation = Get-Location
$projectFilePath = Join-Path $currentLocation "GoProxyWrapper\GoProxyWrapper.csproj"

$msbuildPath = Find-MsBuild

# It'd be simpler to check for $IsWindows, but this check guarantees that this will work properly on old versions of Powershell.
if($IsMacOS -Or $IsLinux) {
    $_ = which dotnet
    $hasNetCli = ($LastExitCode -ne 0)
} else {
    $_ = where.exe dotnet
    $hasNetCli = ($LastExitCode -ne 0)
}

[xml] $proj = Get-Content $projectFilePath

$packageIdNode = Get-XmlNode -XmlDocument $proj -NodePath "Project.PropertyGroup.PackageId"

if($packageIdNode -eq $null) {
    $packageId = $proj.Project.PropertyGroup.Title
} else {
    $packageId = $packageIdNode.InnerText.Trim()
}

$version = FirstObject -Array $proj.Project.PropertyGroup.Version

cd goproxy
if($IsMacOS) {
	bash build.sh
} else {
	./build.bat
}

cd ..

if ($hasNetCli) {
    & dotnet restore $projectFilePath
}

if (-not (Test-Path nuget-packages)) {
    mkdir nuget-packages
}

& .\nuget pack goproxy-native-windows/CloudVeil.goproxy-native-windows.nuspec -OutputDirectory nuget-packages

if($IsMacOS) {
    & nuget pack goproxy-native-macos/CloudVeil.proxy-native-macos.nuspec -OutputDirectory nuget-packages
}

& $msbuildPath /property:Configuration=Release /t:Clean,Build $projectFilePath
& $msbuildPath /property:Configuration=Release /t:pack $projectFilePath

if (-not $?) {
    Write-Host "Error occurred during msbuild process"
    exit 1
}

$nupkg = Join-Path (Split-Path -Path $projectFilePath) "bin/Release/$packageId.$version.nupkg"
$windowsPackage = "nuget-packages/CloudVeil.goproxy-native-windows.$version.nupkg"

if($IsMacOS) {
    $macosPackage = "nuget-packages/CloudVeil.proxy-native-macos.$version.nupkg"
}

if($IsWindows) {
	$destinationPath = "C:\\Nuget.Local"
} else {
	$destinationPath = "~/nuget.local"
}

Copy-Item -Path $nupkg -Destination $destinationPath
Copy-Item -Path $windowsPackage -Destination $destinationPath

if($IsMacOS) {
    Copy-Item -Path $macosPackage -Destination $destinationPath
}

$nugetKeyPath = Join-Path $currentLocation ".nuget-apikey"

if(!(Test-Path $nugetKeyPath)) {
    if($IsWindows) {
		$nugetKeyPath = "C:\\.nuget-apikey"
	} else {
		$nugetKeyPath = "~/.nuget-apikey"
	}
}

if(!(Test-Path $nugetKeyPath)) {
    Write-Host $nugetKeyPath
    Write-Host "Please put .nuget-apikey in your project directory or in C:\ drive root in order to upload to nuget."
    exit
}

$apiKey = Get-Content $nugetKeyPath

$deployNuget = Read-Host -Prompt "Do you want to upload $packageId to nuget?"

if($deployNuget[0] -eq 'y') {
    dotnet nuget push $nupkg -k $apiKey -s https://api.nuget.org/v3/index.json
	dotnet nuget push $windowsPackage -k $apiKey -s https://api.nuget.org/v3/index.json

    if($IsMacOS) {
	    dotnet nuget push $macosPackage -k $apiKey -s https://api.nuget.org/v3/index.json
    }
}
