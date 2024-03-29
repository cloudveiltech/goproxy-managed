﻿using System;
using System.Collections.Generic;
using System.Text;
using System.Runtime.InteropServices;
using System.Security;

namespace GoproxyWrapper
{
    class Const
    {
        /// <summary>
        /// If we have just 'proxy' here, it will allow us to use proxy.dll or libproxy.so for both Windows and macOS.
        /// </summary>
        public const string DLL_PATH = "proxy";
    }

    [StructLayout(LayoutKind.Sequential, CharSet = CharSet.Unicode)]
    struct GoString
    {
        public static GoString FromString(string value)
        {
            GoString s = new GoString();

            if(value == null) {
                s.AsString = "";
            } else {
                s.AsString = value;
            }
            return s;
        }

        public static GoString FromBytes(byte[] value)
        {
            GoString s = new GoString();
            s.AsBytes = value;
            return s;
        }

        IntPtr data;
        public string AsString
        {
            get
            {
                return Marshal.PtrToStringAnsi(data, lengthAsInt);
            }
            set
            {
                data = Marshal.StringToHGlobalAnsi(value);
                lengthAsInt = value.Length;
            }
        }

        // Using IntPtr for length because using Int32 in a 64-bit environment causes length corruption.
        public IntPtr length;

        public int lengthAsInt
        {
            get
            {
                return length.ToInt32();
            }

            set
            {
                length = new IntPtr(value);
            }
        }

        public byte[] AsBytes
        {
            get
            {
                byte[] buffer = new byte[lengthAsInt];
                Marshal.Copy(data, buffer, 0, lengthAsInt);
                return buffer;
            }

            set
            {
                data = Marshal.AllocHGlobal(value.Length);
                Marshal.Copy(value, 0, data, value.Length);

                lengthAsInt = value.Length;
            }
        }
    }

    [StructLayout(LayoutKind.Sequential, CharSet = CharSet.Unicode)]
    struct GoSlice
    {
        IntPtr data;
        public byte[] bytes
        {
            get
            {
                int length = this.length.ToInt32();
                int cap = this.cap.ToInt32();

                byte[] managedArray = new byte[length];
                Marshal.Copy(data, managedArray, 0, length);
                return managedArray;
            }
        }

        // A good way to handle the 4 byte -- 8 byte difference between x64 and x86.
        public IntPtr length;
        public IntPtr cap;
    }

    [SuppressUnmanagedCodeSecurity]
    class ProxyNativeWrapper
    {
        [UnmanagedFunctionPointer(CallingConvention.Cdecl)]
        public delegate int CallbackDelegate(long handle);

        [DllImport(Const.DLL_PATH, CharSet = CharSet.Unicode, CallingConvention = CallingConvention.Cdecl)]
        public static extern void Init(short httpPortNumber, short httpsPortNumber, GoString certFile, GoString keyFile);

        public static void SetProxyLogFile(string logFile)
        {
            GoString str = GoString.FromString(logFile);
            SetProxyLogFile(str);
        }

        [DllImport(Const.DLL_PATH, CharSet = CharSet.Unicode, CallingConvention = CallingConvention.Cdecl)]
        public static extern void SetProxyLogFile(GoString logFile);

        [DllImport(Const.DLL_PATH, CharSet = CharSet.Unicode, CallingConvention = CallingConvention.Cdecl)]
        public static extern void Start();

        [DllImport(Const.DLL_PATH, CharSet = CharSet.Unicode, CallingConvention = CallingConvention.Cdecl)]
        public static extern void Stop();

        [DllImport(Const.DLL_PATH, CharSet = CharSet.Unicode, CallingConvention = CallingConvention.Cdecl)]
        public static extern bool IsRunning();

        [DllImport(Const.DLL_PATH, CharSet = CharSet.Unicode, CallingConvention = CallingConvention.Cdecl)]
        public static extern void GetCert(out GoSlice data);

        [DllImport(Const.DLL_PATH, CharSet = CharSet.Unicode, CallingConvention = CallingConvention.Cdecl)]
        public static extern void SetOnBeforeRequestCallback(CallbackDelegate func);

        [DllImport(Const.DLL_PATH, CharSet = CharSet.Unicode, CallingConvention = CallingConvention.Cdecl)]
        public static extern void SetOnBeforeResponseCallback(CallbackDelegate func);

        [DllImport(Const.DLL_PATH, CharSet = CharSet.Unicode, CallingConvention = CallingConvention.Cdecl)]
        public static extern void SetDestPortForLocalPort(int localPort, int destPort, GoString ip);
    }

    [SuppressUnmanagedCodeSecurity]
    class RequestNativeWrapper
    {
        [DllImport(Const.DLL_PATH, CharSet = CharSet.Unicode, CallingConvention = CallingConvention.Cdecl)]
        public static extern bool RequestGetUrl(long handle, out GoString res);
                
        [DllImport(Const.DLL_PATH, CharSet = CharSet.Unicode, CallingConvention = CallingConvention.Cdecl)]
        public static extern bool RequestGetFirstHeader(long handle, GoString name, out GoString res);

        [DllImport(Const.DLL_PATH, CharSet = CharSet.Unicode, CallingConvention = CallingConvention.Cdecl)]
        public static extern bool RequestHeaderExists(long handle, GoString name);

        [DllImport(Const.DLL_PATH, CharSet = CharSet.Unicode, CallingConvention = CallingConvention.Cdecl)]
        public static extern bool RequestGetBody(long handle, out GoSlice data);

        [DllImport(Const.DLL_PATH, CharSet = CharSet.Unicode, CallingConvention = CallingConvention.Cdecl)]
        public static extern bool RequestGetBodyAsString(long handle, out GoString body);

        [DllImport(Const.DLL_PATH, CharSet = CharSet.Unicode, CallingConvention = CallingConvention.Cdecl)]
        public static extern bool RequestHasBody(long handle);

        [DllImport(Const.DLL_PATH, CharSet = CharSet.Unicode, CallingConvention = CallingConvention.Cdecl)]
        public static extern bool RequestSetHeader(long handle, string name, string value);

        [DllImport(Const.DLL_PATH, CharSet = CharSet.Unicode, CallingConvention = CallingConvention.Cdecl)]
        public static extern int RequestGetHeaders(long handle, out GoString res);
    }

    [SuppressUnmanagedCodeSecurity]
    class ResponsetNativeWrapper
    {
        [DllImport(Const.DLL_PATH, CharSet = CharSet.Unicode, CallingConvention = CallingConvention.Cdecl)]
        public static extern bool ResponseGetFirstHeader(long handle, GoString name, out GoString res);

        [DllImport(Const.DLL_PATH, CharSet = CharSet.Unicode, CallingConvention = CallingConvention.Cdecl)]
        public static extern bool ResponseHeaderExists(long handle, GoString name);

        [DllImport(Const.DLL_PATH, CharSet = CharSet.Unicode, CallingConvention = CallingConvention.Cdecl)]
        public static extern bool ResponseGetBody(long handle, out GoSlice data);

        [DllImport(Const.DLL_PATH, CharSet = CharSet.Unicode, CallingConvention = CallingConvention.Cdecl)]
        public static extern bool ResponseGetBodyAsString(long handle, out GoString body);

        [DllImport(Const.DLL_PATH, CharSet = CharSet.Unicode, CallingConvention = CallingConvention.Cdecl)]
        public static extern bool ResponseHasBody(long handle);

        [DllImport(Const.DLL_PATH, CharSet = CharSet.Unicode, CallingConvention = CallingConvention.Cdecl)]
        public static extern bool ResponseSetHeader(long handle, string name, string value);

        [DllImport(Const.DLL_PATH, CharSet = CharSet.Unicode, CallingConvention = CallingConvention.Cdecl)]
        public static extern bool CreateResponse(long handle, int status, GoString contentType, GoString body);

        [DllImport(Const.DLL_PATH, CharSet = CharSet.Unicode, CallingConvention = CallingConvention.Cdecl)]
        public static extern int ResponseGetHeaders(long handle, out GoString res);

        [DllImport(Const.DLL_PATH, CharSet = CharSet.Unicode, CallingConvention = CallingConvention.Cdecl)]
        public static extern int ResponseGetCertificatesCount(long handle);

        [DllImport(Const.DLL_PATH, CharSet = CharSet.Unicode, CallingConvention = CallingConvention.Cdecl)]
        public static extern int ResponseGetCertificate(long handle, int index, out GoSlice certBytes);

        [DllImport(Const.DLL_PATH, CharSet = CharSet.Unicode, CallingConvention = CallingConvention.Cdecl)]
        public static extern bool ResponseIsTLSVerified(long handle);
    }
}

