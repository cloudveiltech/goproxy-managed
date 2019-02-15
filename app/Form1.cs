using System;
using GoproxyWrapper;
using System.Windows.Forms;
using System.Timers;
using System.Text;
using System.Diagnostics;

namespace testapp
{
    public partial class Form1 : Form
    {
        public Form1()
        {
            InitializeComponent();

            AppDomain currentDomain = AppDomain.CurrentDomain;
            currentDomain.UnhandledException += new UnhandledExceptionEventHandler(MyHandler);

            timer.Start();
        }

        static void MyHandler(object sender, UnhandledExceptionEventArgs args)
        {
            Exception e = (Exception)args.ExceptionObject;
            Console.WriteLine("MyHandler caught : " + e.Message);
            Console.WriteLine("Runtime terminating: {0}", args.IsTerminating);
        }

        private void init_Click(object sender, EventArgs e)
        {
            short portHttp = 0;
            short portHttps = 0;
            if (short.TryParse(portNumberHttp.Text, out portHttp) && short.TryParse(portNumberHttps.Text, out portHttps))
            {
                GoProxy.Instance.Init(portHttp, portHttps, "rootCertificate.pem", "rootPrivateKey.pem");
                GoProxy.Instance.BeforeRequest += Instance_BeforeRequest;
                GoProxy.Instance.BeforeResponse += Instance_BeforeResponse;
                AppendLog("Initialized \r\n");
            }
        }

        private void AppendLog(string log)
        {
            Debug.WriteLine(log);
            logView.Invoke((MethodInvoker)delegate
            {
                logView.AppendText(log);
            });
        }

        private void Instance_BeforeResponse(Session session)
        {
        //    AppendLog("Response " + session.Request.Url + "\r\n");

            foreach(var cert in session.Response.Certificates)
            {
                //   AppendLog("Certificate https: ");
                //   AppendLog(info.Hash);
                //   AppendLog("\r\n Domains: ");
                AppendLog(cert.GetNameInfo(System.Security.Cryptography.X509Certificates.X509NameType.DnsFromAlternativeName, false));
                AppendLog("\r\n");
                AppendLog(cert.GetNameInfo(System.Security.Cryptography.X509Certificates.X509NameType.EmailName, false));
                AppendLog("\r\n");
                //   AppendLog("\r\n");
            }
        /*    foreach (Header h in session.Response.Headers)
            {
                AppendLog(h.RawValue + "\r\n");
            }

            
                        /**Response headers*/
            // var exists = session.Response.Headers.IsHeaderExist("Content-Type");
            // AppendLog("Content-Type exists: " + exists + "\r\n");

     /*       var contentType = session.Response.Headers.GetFirstHeader("Content-Type");
            if (contentType != null)
            {
                if (contentType.Value.Contains("mp4") || contentType.Value.Contains("image"))
                {
                    session.SendCustomResponse(Session.FORBIDDEN, contentType.Value, "This is blocked by GoProxy Wrapper");
                    AppendLog("Blocking mp4 and image\r\n");
                    return;
                }
                AppendLog(contentType.RawValue + "\r\n");

                //**Response body**
                var exists = session.Response.HasBody;
                AppendLog("Has body: " + exists + "\r\n");
                if (exists && contentType.Value.ToLower().Contains("text"))
                {
                    var body = session.Response.BodyAsString;
                   
                    if(body.Contains("Anasayfa"))
                    {
                        AppendLog("Body inspection blocking\r\n");
                        session.SendCustomResponse(Session.FORBIDDEN, Session.CONTENT_TYPE_HTML, "<b>Anasayfa</b> is blocked by GoProxy Wrapper");
                    }
                }
            }*/
        }

        private void Instance_BeforeRequest(Session session)
        {
         //   AppendLog("Request " + session.Request.Url + "\r\n");

            //      **Request headers**
  /*          foreach (Header h in session.Request.Headers)
            {
                AppendLog(h.RawValue + "\r\n");
            }
*/
            /*     bool exists = session.Request.Headers.IsHeaderExist("X-TEST-HEADER");
                  AppendLog("X-TEST-HEADER exists: " + exists + "\r\n");
            /*
                  exists = session.Request.Headers.IsHeaderExist("Accept-Language");
                  AppendLog("Accept-Language exists: " + exists + "\r\n");
                  if(exists)
                  {
                      var v = session.Request.Headers.GetFirstHeader("Accept-Language");
                      AppendLog(v.RawValue + "\r\n");
                  }
                  */


            //**Request body**
            /* bool exists = session.Request.HasBody;
             AppendLog("Has body: " + exists + "\r\n");
             if(exists)
             {
                 AppendLog("Body: " + + "\r\n");
             }
             */


            /**custom response*/

            if (session.Request.Url.Contains("forum"))
            {
                session.SendCustomResponse(Session.FORBIDDEN, Session.CONTENT_TYPE_HTML, "<b>This page is blocked by GoProxy Wrapper</b>");
            }
            
        }

        private void start_Click(object sender, EventArgs e)
        {
            GoProxy.Instance.Start();
            AppendLog("Started \r\n");
        }

        private void stop_Click(object sender, EventArgs e)
        {
            GoProxy.Instance.Stop();
            AppendLog("Stopped \r\n");
        }

        private void timer_Tick(object sender, EventArgs e)
        {
            runningLabel.Text = GoProxy.Instance.IsRunning ? "Running: true" : "Running: false";
            start.Enabled = !GoProxy.Instance.IsRunning;
            stop.Enabled = GoProxy.Instance.IsRunning;
        }
    }
}
