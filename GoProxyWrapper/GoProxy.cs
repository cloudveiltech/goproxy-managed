using GoProxyWrapper;
using System;
using System.Diagnostics;
using System.Runtime.InteropServices;

namespace GoproxyWrapper
{
    public sealed class GoProxy
    {
        static GoProxy()
        {
            
        }

        private GoProxy() { }

        public static GoProxy Instance { get; } = new GoProxy();

        private ProxyNativeWrapper.CallbackDelegate onBeforeRequestDelegate;
        private ProxyNativeWrapper.CallbackDelegate onBeforeResponseDelegate;

        private AdBlockMatcherApi.InternalAdBlockCallbackDelegate onBlacklistDelegate;
        private AdBlockMatcherApi.InternalAdBlockCallbackDelegate onWhitelistDelegate;

        /// <summary>
        /// Calls goproxy initalization and loads certificate and key from the specified files.
        /// </summary>
        /// <param name="portNumber"></param>
        /// <param name="certFile">string containing a path to a PEM-encoded certificate file.</param>
        /// <param name="keyFile">string containing a path to a PEM-encoded key file.</param>
        public void Init(short httpPortNumber, short httpsPortNumber, string certFile, string keyFile)
        {
            onBeforeRequestDelegate = new ProxyNativeWrapper.CallbackDelegate(onBeforeRequest);
            onBeforeResponseDelegate = new ProxyNativeWrapper.CallbackDelegate(onBeforeResponse);

            onBlacklistDelegate = new AdBlockMatcherApi.InternalAdBlockCallbackDelegate(onBlacklist);
            onWhitelistDelegate = new AdBlockMatcherApi.InternalAdBlockCallbackDelegate(onWhitelist);

            ProxyNativeWrapper.SetOnBeforeRequestCallback(onBeforeRequestDelegate);
            ProxyNativeWrapper.SetOnBeforeResponseCallback(onBeforeResponseDelegate);

            AdBlockMatcherApi.SetBlacklistCallback(onBlacklistDelegate);
            AdBlockMatcherApi.SetWhitelistCallback(onWhitelistDelegate);

            ProxyNativeWrapper.Init(httpPortNumber, httpsPortNumber, GoString.FromString(certFile), GoString.FromString(keyFile));
        }

        private int onWhitelist(long handle, GoString goUrl, IntPtr categoriesPtr, int categoryLen)
        {
            int[] categories = new int[categoryLen];

            Marshal.Copy(categoriesPtr, categories, 0, categoryLen);
            Session session = new Session(handle, new Request(handle), new Response(handle));
            string url = goUrl.AsString;

            return Whitelisted?.Invoke(session, url, categories) ?? 0;
        }

        private int onBlacklist(long handle, GoString goUrl, IntPtr categoriesPtr, int categoryLen)
        {
            int[] categories = new int[categoryLen];

            Marshal.Copy(categoriesPtr, categories, 0, categoryLen);
            Session session = new Session(handle, new Request(handle), new Response(handle));
            string url = goUrl.AsString;

            return Blacklisted?.Invoke(session, url, categories) ?? 0;
        }

        private int onBeforeRequest(long handle)
        {
            return (int)BeforeRequest?.Invoke(new Session(handle, new Request(handle), new Response(handle)));
        }

        private int onBeforeResponse(long handle)
        {
            BeforeResponse?.Invoke(new Session(handle, new Request(handle), new Response(handle)));

            return 0;
        }

        public void Start()
        {
            ProxyNativeWrapper.Start();
        }

        public void Stop()
        {
            ProxyNativeWrapper.Stop();
        }

        public void SetProxyLogFile(string logFile)
        {
            ProxyNativeWrapper.SetProxyLogFile(logFile);
        }

        public bool IsRunning
        {
            get { return ProxyNativeWrapper.IsRunning(); }
        }

        public byte[] CertificateBytes
        {
           get
            {
                GoSlice slice = new GoSlice();
                ProxyNativeWrapper.GetCert(out slice);
                return slice.bytes;
            }
        }

        public delegate int OnBeforeRequest(Session session);
        public delegate void OnBeforeResponse(Session session);

        public event OnBeforeRequest BeforeRequest;
        public event OnBeforeResponse BeforeResponse;

        public event AdBlockMatcherApi.AdBlockCallbackDelegate Blacklisted;
        public event AdBlockMatcherApi.AdBlockCallbackDelegate Whitelisted;
    }
}
