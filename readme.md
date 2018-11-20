# GoProxy wrapper
This is go proxy wrapped for .NET 2.0. Repo contains golang part and c# part. 

-  .NET MS Visual Studio debugger conflicts with GO Crash Handler. 
So in order to work it normally you have to apply a small patch to go runtime.
In file `c:\Go\src\runtime\signal_windows.go`, line `154` or whatever path you have add
 ```golang
 // lastcontinuehandler is reached, because runtime cannot handle
// current exception. lastcontinuehandler will print crash info and exit.
//
// It is nosplit for the same reason as exceptionhandler.
//
//go:nosplit
func lastcontinuehandler(info *exceptionrecord, r *context, gp *g) int32 {
	if testingWER {
		return _EXCEPTION_CONTINUE_SEARCH
	}

	return _EXCEPTION_CONTINUE_SEARCH //this is new line. We had to pass all of this to .NET handler so debugging would work in MSVS
 ```

- Custom certificate should be placed in the same directory as dll file with names `cert.pem` and `key.pem`
- C header with all APIs you can see in `proxy.h` file
- To build this library just run `build.bat`
- Compiled dll is placed to `testapp/proxy.dll`
- All interaction samples are place into `Form1.cs`
- Path to Dll can be altered in `GoproxyWrapper.Const` class
- All wrapper classes is placed into `app/GoproxyWrapper`