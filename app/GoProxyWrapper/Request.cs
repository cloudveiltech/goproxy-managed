using System.Collections;
using System.Collections.Generic;

namespace GoproxyWrapper
{
    public sealed class Request
    {
        private long handle;
        private string url = "";
        private byte[] body;

        public Request(long handle)
        {
            this.handle = handle;
            Headers = new HeaderCollection(handle);
        }

        public HeaderCollection Headers { get; }

        public string Url
        {
            get
            {
                if (url.Length > 0)
                {
                    return url;
                }
                GoString str = new GoString();
                if (RequestNativeWrapper.RequestGetUrl(handle, out str))
                {
                    url = str.AsString;
                    return url;
                }

                return "";
            }
        }

        public bool HasBody
        {
            get
            {
                return RequestNativeWrapper.RequestHasBody(handle);
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
                if (RequestNativeWrapper.RequestGetBody(handle, out slice))
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
                if (RequestNativeWrapper.RequestGetBodyAsString(handle, out bodyString))
                {
                    return bodyString.AsString;
                }
                return "";
            }
        }

        public class HeaderCollection: IEnumerable<Header>
        {
            private long requestHandle;
            List<Header> marshalledHeaders;

            public HeaderCollection(long requestHandle)
            {
                this.requestHandle = requestHandle;
            }

         
            public Header GetFirstHeader(string name)
            {
                GoString res = new GoString();
                if (RequestNativeWrapper.RequestGetFirstHeader(requestHandle, GoString.FromString(name), out res))
                {
                    return new Header(name, res.AsString);
                }
                else
                {
                    return null;
                }
            }

            private List<Header> Headers
            {
                get
                {
                    if(marshalledHeaders != null)
                    {
                        return marshalledHeaders;
                    }

                    GoString s = new GoString();
                    marshalledHeaders = new List<Header>();
                    if(RequestNativeWrapper.RequestGetHeaders(requestHandle, out s) == 0)
                    {
                        return marshalledHeaders;
                    }

                    string[] headers = s.AsString.Split(new string[] { "\r\n" }, System.StringSplitOptions.RemoveEmptyEntries);

                    for(int i=0; i<headers.Length; i++)
                    {
                        marshalledHeaders.Add(new Header(headers[i]));
                    }

                    return marshalledHeaders;
                }
            }

            public bool HeaderExists(string name)
            {
                return RequestNativeWrapper.RequestHeaderExists(requestHandle, GoString.FromString(name));
            }

            public void SetHeader(string name, string value)
            {
                RequestNativeWrapper.RequestSetHeader(requestHandle, name, value);
                marshalledHeaders = null;
            }

            public void SetHeader(Header header)
            {
                RequestNativeWrapper.RequestSetHeader(requestHandle, header.Name, header.Value);
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
