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
            get;
            private set;
        }

        public string Value
        {
            get;
            private set;
        }

        public string RawValue
        {
            get
            {
                return Name + ": " + Value;
            }

            set
            {
                var parts = value.Split(new char[] { ':'});
                Name = parts[0];
                if (parts.Length > 1)
                {
                    Value = parts[1];
                }
            }
        }
    }
}
