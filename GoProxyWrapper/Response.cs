using System.Collections;
using System.Collections.Generic;
using System.Security.Cryptography.X509Certificates;

namespace GoproxyWrapper
{
    public sealed class Response
    {
        private long handle;
        private byte[] body;
        private List<X509Certificate2> cachedCertifcates;


        public Response(long handle)
        {
            this.handle = handle;
            Headers = new HeaderCollection(handle);
        }

        public HeaderCollection Headers { get; }

        public bool IsCertificateVerified
            => ResponsetNativeWrapper.ResponseIsTLSVerified(this.handle);

        public int CertificateCount => ResponsetNativeWrapper.ResponseGetCertificatesCount(this.handle);

        public List<X509Certificate2> Certificates
        {
            get
            {
                if (cachedCertifcates != null)
                {
                    return cachedCertifcates;
                }

                cachedCertifcates = new List<X509Certificate2>();

                int certificateCount = ResponsetNativeWrapper.ResponseGetCertificatesCount(handle);
                if (certificateCount == 0)
                {
                    return cachedCertifcates;
                }

                for (int i = 0; i < certificateCount; i++)
                {
                    GoSlice certData = new GoSlice();

                    ResponsetNativeWrapper.ResponseGetCertificate(handle, i, out certData);

                    X509Certificate2 cert = new X509Certificate2(certData.bytes);
                    
                    cachedCertifcates.Add(cert);
                }
                return cachedCertifcates;
            }

        }

        public bool HasBody
        {
            get
            {
                return ResponsetNativeWrapper.ResponseHasBody(handle);
            }
        }

        public byte[] Body
        {
            get
            {
                if (body != null || !HasBody)
                {
                    return body;
                }
                GoSlice slice = new GoSlice();
                if (ResponsetNativeWrapper.ResponseGetBody(handle, out slice))
                {
                    body = slice.bytes;
                    return body;
                }
                return null;
            }
        }

        public string BodyAsString
        {
            get
            {
                GoString bodyString = new GoString();
                if(ResponsetNativeWrapper.ResponseGetBodyAsString(handle, out bodyString))
                {
                    return bodyString.AsString;
                }
                return "";
            }
        }

        public class HeaderCollection : IEnumerable<Header>
        {
            private long responseHandle;
            private List<Header> marshalledHeaders;

            public HeaderCollection(long responseHandle)
            {
                this.responseHandle = responseHandle;
            }

            private List<Header> Headers
            {
                get
                {
                    if (marshalledHeaders != null)
                    {
                        return marshalledHeaders;
                    }

                    GoString s = new GoString();
                    marshalledHeaders = new List<Header>();

                    if(ResponsetNativeWrapper.ResponseGetHeaders(responseHandle, out s) == 0)
                    {
                        return marshalledHeaders;
                    }

                    string[] headers = s.AsString.Split(new string[] { "\r\n" }, System.StringSplitOptions.RemoveEmptyEntries);

                    for (int i = 0; i < headers.Length; i++)
                    {
                        marshalledHeaders.Add(new Header(headers[i]));
                    }

                    return marshalledHeaders;
                }
            }

            public Header GetFirstHeader(string name)
            {
                GoString res = new GoString();
                if (ResponsetNativeWrapper.ResponseGetFirstHeader(responseHandle, GoString.FromString(name), out res))
                {
                    return new Header(name, res.AsString);
                }
                else
                {
                    return null;
                }
            }

            public bool HeaderExists(string name)
            {
                return ResponsetNativeWrapper.ResponseHeaderExists(responseHandle, GoString.FromString(name));
            }

            public void SetHeader(string name, string value)
            {
                ResponsetNativeWrapper.ResponseSetHeader(responseHandle, name, value);
                marshalledHeaders = null;
            }

            public void SetHeader(Header header)
            {
                ResponsetNativeWrapper.ResponseSetHeader(responseHandle, header.Name, header.Value);
                marshalledHeaders = null;
            }

            public IEnumerator<Header> GetEnumerator()
            {
                return Headers.GetEnumerator();
            }

            IEnumerator IEnumerable.GetEnumerator()
            {
                return GetEnumerator();
            }
        }
    }
}
