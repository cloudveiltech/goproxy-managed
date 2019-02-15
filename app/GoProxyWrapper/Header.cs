namespace GoproxyWrapper
{
    public sealed class Header
    {
        public Header(string name, string value)
        {
            Name = name;
            Value = value;
        }

        public Header(string rawValue)
        {
            RawValue = rawValue;
        }

        public string Name
        {
            get => headerParts[0];
            private set => updateHeaderPart(0, value);
        }

        public string Value
        {
            get => headerParts[1];
            private set => updateHeaderPart(1, value);
        }

        private string rawValue;
        private string[] headerParts;
        public string RawValue
        {
            get
            {
                return rawValue;
            }

            set
            {
                rawValue = value;
                headerParts = value.Split(new char[] { ':'}, 2);
            }
        }

        private void updateHeaderPart(int part, string value)
        {
            if(headerParts == null || headerParts.Length < 2)
            {
                headerParts = new string[2];
            }

            headerParts[part] = value;
            RawValue = headerParts[0] + ": " + headerParts[1];
        }
    }
}
