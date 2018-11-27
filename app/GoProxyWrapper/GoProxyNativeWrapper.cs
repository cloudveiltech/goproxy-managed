using System;
using System.Collections.Generic;
using System.Text;
using System.Runtime.InteropServices;
using System.Security;

namespace GoproxyWrapper
{
    class Const
    {
        public const string DLL_PATH = @"d:\work\Filter-Windows\GoProxyDotNet\goproxy\bin\x64\proxy.dll";
    }

    [StructLayout(LayoutKind.Sequential, CharSet = CharSet.Unicode)]
    struct GoString
    {
        public static GoString FromString(string value)
        {
            GoString s = new GoString();
            s.AsString = value;
            return s;
        }

        IntPtr data;
        public string AsString
        {
            get
            {
                return Marshal.PtrToStringAnsi(data, length);
            }
            set
            {
                data = Marshal.StringToHGlobalAnsi(value);
                length = value.Length;
            }
        }
        public int length;
    }

    [StructLayout(LayoutKind.Sequential, CharSet = CharSet.Unicode)]
    struct GoSlice
    {
        IntPtr data;
        public byte[] bytes
        {
            get
            {
                byte[] managedArray = new byte[length - cap];
                Marshal.Copy(data, managedArray, cap, length - cap);
                return managedArray;
            }
        }
        public int length;
        public int cap;
    }

    [SuppressUnmanagedCodeSecurity]
    class ProxyNativeWrapper
    {
        [UnmanagedFunctionPointer(CallingConvention.StdCall)]
        public delegate void CallbackDelegate(long handle);


        [DllImport(Const.DLL_PATH, CharSet = CharSet.Unicode, CallingConvention = CallingConvention.StdCall)]
        public static extern void Init(int portNumber);

        [DllImport(Const.DLL_PATH, CharSet = CharSet.Unicode, CallingConvention = CallingConvention.StdCall)]
        public static extern void Start();

        [DllImport(Const.DLL_PATH, CharSet = CharSet.Unicode, CallingConvention = CallingConvention.StdCall)]
        public static extern void Stop();

        [DllImport(Const.DLL_PATH, CharSet = CharSet.Unicode, CallingConvention = CallingConvention.StdCall)]
        public static extern bool IsRunning();

        [DllImport(Const.DLL_PATH, CharSet = CharSet.Unicode, CallingConvention = CallingConvention.StdCall)]
        public static extern void GetCert(out GoSlice data);

        [DllImport(Const.DLL_PATH, CharSet = CharSet.Unicode, CallingConvention = CallingConvention.StdCall)]
        public static extern void SetOnBeforeRequestCallback(CallbackDelegate func);

        [DllImport(Const.DLL_PATH, CharSet = CharSet.Unicode, CallingConvention = CallingConvention.StdCall)]
        public static extern void SetOnBeforeResponseCallback(CallbackDelegate func);
    }

    [SuppressUnmanagedCodeSecurity]
    class RequestNativeWrapper
    {
        [DllImport(Const.DLL_PATH, CharSet = CharSet.Unicode, CallingConvention = CallingConvention.StdCall)]
        public static extern bool RequestGetUrl(long handle, out GoString res);
                
        [DllImport(Const.DLL_PATH, CharSet = CharSet.Unicode, CallingConvention = CallingConvention.StdCall)]
        public static extern bool RequestGetFirstHeader(long handle, GoString name, out GoString res);

        [DllImport(Const.DLL_PATH, CharSet = CharSet.Unicode, CallingConvention = CallingConvention.StdCall)]
        public static extern bool RequestHeaderExists(long handle, GoString name);

        [DllImport(Const.DLL_PATH, CharSet = CharSet.Unicode, CallingConvention = CallingConvention.StdCall)]
        public static extern bool RequestGetBody(long handle, out GoSlice data);

        [DllImport(Const.DLL_PATH, CharSet = CharSet.Unicode, CallingConvention = CallingConvention.StdCall)]
        public static extern bool RequestGetBodyAsString(long handle, out GoString body);

        [DllImport(Const.DLL_PATH, CharSet = CharSet.Unicode, CallingConvention = CallingConvention.StdCall)]
        public static extern bool RequestHasBody(long handle);

        [DllImport(Const.DLL_PATH, CharSet = CharSet.Unicode, CallingConvention = CallingConvention.StdCall)]
        public static extern bool RequestSetHeader(long handle, string name, string value);

        [DllImport(Const.DLL_PATH, CharSet = CharSet.Unicode, CallingConvention = CallingConvention.StdCall)]
        public static extern int RequestGetHeaders(long handle, out GoString res);
    }

    [SuppressUnmanagedCodeSecurity]
    class ResponsetNativeWrapper
    {
        [DllImport(Const.DLL_PATH, CharSet = CharSet.Unicode, CallingConvention = CallingConvention.StdCall)]
        public static extern bool ResponseGetFirstHeader(long handle, GoString name, out GoString res);

        [DllImport(Const.DLL_PATH, CharSet = CharSet.Unicode, CallingConvention = CallingConvention.StdCall)]
        public static extern bool ResponseHeaderExists(long handle, GoString name);

        [DllImport(Const.DLL_PATH, CharSet = CharSet.Unicode, CallingConvention = CallingConvention.StdCall)]
        public static extern bool ResponseGetBody(long handle, out GoSlice data);

        [DllImport(Const.DLL_PATH, CharSet = CharSet.Unicode, CallingConvention = CallingConvention.StdCall)]
        public static extern bool ResponseGetBodyAsString(long handle, out GoString body);

        [DllImport(Const.DLL_PATH, CharSet = CharSet.Unicode, CallingConvention = CallingConvention.StdCall)]
        public static extern bool ResponseHasBody(long handle);

        [DllImport(Const.DLL_PATH, CharSet = CharSet.Unicode, CallingConvention = CallingConvention.StdCall)]
        public static extern bool ResponseSetHeader(long handle, string name, string value);

        [DllImport(Const.DLL_PATH, CharSet = CharSet.Unicode, CallingConvention = CallingConvention.StdCall)]
        public static extern bool CreateResponse(long handle, int status, GoString contentType, GoString body);

        [DllImport(Const.DLL_PATH, CharSet = CharSet.Unicode, CallingConvention = CallingConvention.StdCall)]
        public static extern int ResponseGetHeaders(long handle, out GoString res);
    }
}

