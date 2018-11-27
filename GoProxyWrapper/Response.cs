using System.Collections.Generic;

namespace GoproxyWrapper
{
    public sealed class Response
    {
        private long handle;
        private byte[] body;

        public Response(long handle)
        {
            this.handle = handle;
            Headers = new HeaderCollection(handle);
        }

        public HeaderCollection Headers { get; }

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

        public class HeaderCollection
        {
            private long responseHandle;
            public HeaderCollection(long responseHandle)
            {
                this.responseHandle = responseHandle;
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
            }

            public void SetHeader(Header header)
            {
                ResponsetNativeWrapper.ResponseSetHeader(responseHandle, header.Name, header.Value);
            }
        }
    }
}
