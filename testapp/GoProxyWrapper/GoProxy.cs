using System;
using System.Diagnostics;

namespace GoproxyWrapper
{
    public sealed class GoProxy
    {
        static GoProxy() { }

        private GoProxy() { }

        public static GoProxy Instance { get; } = new GoProxy();

        private ProxyNativeWrapper.CallbackDelegate onBeforeRequestDelegate;
        private ProxyNativeWrapper.CallbackDelegate onBeforeResponseDelegate;

        public void Init(int portNumber)
        {
            onBeforeRequestDelegate = new ProxyNativeWrapper.CallbackDelegate(onBeforeRequest);
            onBeforeResponseDelegate = new ProxyNativeWrapper.CallbackDelegate(onBeforeResponse);

            ProxyNativeWrapper.SetOnBeforeRequestCallback(onBeforeRequestDelegate);
            ProxyNativeWrapper.SetOnBeforeResponseCallback(onBeforeResponseDelegate);
            ProxyNativeWrapper.Init(portNumber);
        }

        private void onBeforeRequest(long handle)
        {
            BeforeRequest?.Invoke(new Session(handle, new Request(handle), new Response(handle)));
        }

        private void onBeforeResponse(long handle)
        {
            BeforeResponse?.Invoke(new Session(handle, new Request(handle), new Response(handle)));
        }

        public void Start()
        {
            ProxyNativeWrapper.Start();
        }

        public void Stop()
        {
            ProxyNativeWrapper.Stop();
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

        public delegate void OnBeforeRequest(Session session);
        public delegate void OnBeforeResponse(Session session);

        public event OnBeforeRequest BeforeRequest;
        public event OnBeforeResponse BeforeResponse;
    }
}
