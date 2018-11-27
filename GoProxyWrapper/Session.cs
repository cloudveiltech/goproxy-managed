namespace GoproxyWrapper
{
    public sealed class Session
    {
        public const string CONTENT_TYPE_HTML = "text/html";
        public const string CONTENT_TYPE_TEXT = "text/plain";
        public const int FORBIDDEN = 403;        

        private long handle;

        public Session(long handle, Request request, Response response)
        {
            this.handle = handle;
            Request = request;
            Response = response;
        }

        public Request Request
        {
            get;
        }

        public Response Response
        {
            get;
        }

        public bool SendCustomResponse(int responseCode, string contentType, string body)
        {
            return ResponsetNativeWrapper.CreateResponse(handle, responseCode, GoString.FromString(contentType), GoString.FromString(body));
        }

        public bool SendCustomResponse(int responseCode, string contentType, byte[] body)
        {
            return ResponsetNativeWrapper.CreateResponse(handle, responseCode, GoString.FromString(contentType), GoString.FromBytes(body));
        }
    }
}
