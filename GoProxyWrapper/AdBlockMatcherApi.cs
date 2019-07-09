﻿using GoproxyWrapper;
using System;
using System.Collections.Generic;
using System.Runtime.InteropServices;
using System.Text;

namespace GoProxyWrapper
{
    public static class AdBlockMatcherApi
    {
        [DllImport(Const.DLL_PATH, CharSet = CharSet.Unicode, CallingConvention = CallingConvention.Cdecl, EntryPoint = "AdBlockMatcherInitialize")]
        public static extern void Initialize();

        [DllImport(Const.DLL_PATH, CharSet = CharSet.Unicode, CallingConvention = CallingConvention.Cdecl, EntryPoint = "AdBlockMatcherParseRuleFile")]
        internal static extern void ParseRuleFile(GoString fileName, int categoryId, ListType listType);

        public static void ParseRuleFile(string fileName, int categoryId, ListType listType)
        {
            GoString gsFileName = GoString.FromString(fileName);
            ParseRuleFile(gsFileName, categoryId, listType);
        }

        [DllImport(Const.DLL_PATH, CharSet = CharSet.Unicode, CallingConvention = CallingConvention.Cdecl, EntryPoint = "AdBlockMatcherLoad")]
        internal static extern void Load(GoString fileName);

        public static void Load(string fileName)
        {
            GoString gsFileName = GoString.FromString(fileName);
            Load(gsFileName);
        }

        [DllImport(Const.DLL_PATH, CharSet = CharSet.Unicode, CallingConvention = CallingConvention.Cdecl, EntryPoint = "AdBlockMatcherSave")]
        internal static extern void Save(GoString filePath);

        public static void Save(string filePath)
        {
            GoString gsFilePath = GoString.FromString(filePath);
            Save(gsFilePath);
        }

        [DllImport(Const.DLL_PATH, CharSet = CharSet.Unicode, CallingConvention = CallingConvention.Cdecl, EntryPoint = "AdBlockMatcherTestUrlMatch")]
        internal static extern int TestUrlMatch(GoString url, GoString host);

        public static int TestUrlMatch(string url, string host)
        {
            GoString gsUrl = GoString.FromString(url);
            GoString gsHost = GoString.FromString(host);

            return TestUrlMatch(gsUrl, gsHost);
        }

        [DllImport(Const.DLL_PATH, CharSet = CharSet.Unicode, CallingConvention = CallingConvention.Cdecl, EntryPoint = "AdBlockMatcherAreListsLoaded")]
        public static extern bool AreListsLoaded();
    }
}
