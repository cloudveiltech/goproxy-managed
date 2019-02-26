# Building

## Windows

### Step 1. Install build tools
The first thing you need is the Golang runtime and tools. Download them here and install them: https://golang.org/dl/

**NOTE: Might need to expand on 32-bit/64-bit explanation a bit**

The next thing you need is MSYS2: https://www.msys2.org

**NOTE:** `build.bat` assumes that MSYS2 is installed at `C:\msys64`

Once it's installed, open the MSYS2 shell and run

```
pacman -Syuu
```

You might need to close the terminal and reopen, but just run this until there are no more updates.

The next thing you need to run is

```
pacman -S mingw-w64-x86_64-toolchain mingw-w64-i686-toolchain
```

This installs the proper cross-compilers for Windows 32-bit and 64-bit

### Step 2. Build

Go to the root level of the repository and run

```
.\nuget-deploy.ps1
```

This will generate DLLs and a nuget package automatically.

**NOTE:** When the build script asks you if you want to upload to nuget, just type `n`. If you want to auto-upload to nuget, you'll need to obtain an API key from nuget.org

