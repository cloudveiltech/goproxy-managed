using System;

namespace GoproxyWrapper
{
	public enum ProxyNextAction
	{
		AllowAndIgnoreContent = 0,
		AllowButRequestContentInspection = 1,
		AllowAndIgnoreContentAndResponse = 2,
		DropConnection = 3
	}
}
