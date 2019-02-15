namespace testapp
{
    partial class Form1
    {
        /// <summary>
        /// Required designer variable.
        /// </summary>
        private System.ComponentModel.IContainer components = null;

        /// <summary>
        /// Clean up any resources being used.
        /// </summary>
        /// <param name="disposing">true if managed resources should be disposed; otherwise, false.</param>
        protected override void Dispose(bool disposing)
        {
            if (disposing && (components != null))
            {
                components.Dispose();
            }
            base.Dispose(disposing);
        }

        #region Windows Form Designer generated code

        /// <summary>
        /// Required method for Designer support - do not modify
        /// the contents of this method with the code editor.
        /// </summary>
        private void InitializeComponent()
        {
            this.components = new System.ComponentModel.Container();
            this.start = new System.Windows.Forms.Button();
            this.stop = new System.Windows.Forms.Button();
            this.init = new System.Windows.Forms.Button();
            this.portNumberHttp = new System.Windows.Forms.TextBox();
            this.label1 = new System.Windows.Forms.Label();
            this.runningLabel = new System.Windows.Forms.Label();
            this.timer = new System.Windows.Forms.Timer(this.components);
            this.logView = new System.Windows.Forms.TextBox();
            this.portNumberHttps = new System.Windows.Forms.TextBox();
            this.label2 = new System.Windows.Forms.Label();
            this.label3 = new System.Windows.Forms.Label();
            this.SuspendLayout();
            // 
            // start
            // 
            this.start.Location = new System.Drawing.Point(59, 115);
            this.start.Name = "start";
            this.start.Size = new System.Drawing.Size(75, 23);
            this.start.TabIndex = 0;
            this.start.Text = "Start";
            this.start.UseVisualStyleBackColor = true;
            this.start.Click += new System.EventHandler(this.start_Click);
            // 
            // stop
            // 
            this.stop.Enabled = false;
            this.stop.Location = new System.Drawing.Point(59, 144);
            this.stop.Name = "stop";
            this.stop.Size = new System.Drawing.Size(75, 23);
            this.stop.TabIndex = 1;
            this.stop.Text = "Stop";
            this.stop.UseVisualStyleBackColor = true;
            this.stop.Click += new System.EventHandler(this.stop_Click);
            // 
            // init
            // 
            this.init.Location = new System.Drawing.Point(59, 86);
            this.init.Name = "init";
            this.init.Size = new System.Drawing.Size(75, 23);
            this.init.TabIndex = 2;
            this.init.Text = "Init";
            this.init.UseVisualStyleBackColor = true;
            this.init.Click += new System.EventHandler(this.init_Click);
            // 
            // portNumberHttp
            // 
            this.portNumberHttp.Location = new System.Drawing.Point(49, 33);
            this.portNumberHttp.MaxLength = 5;
            this.portNumberHttp.Name = "portNumberHttp";
            this.portNumberHttp.Size = new System.Drawing.Size(100, 20);
            this.portNumberHttp.TabIndex = 3;
            this.portNumberHttp.Text = "8081";
            this.portNumberHttp.WordWrap = false;
            // 
            // label1
            // 
            this.label1.AutoSize = true;
            this.label1.Location = new System.Drawing.Point(68, 17);
            this.label1.Name = "label1";
            this.label1.Size = new System.Drawing.Size(66, 13);
            this.label1.TabIndex = 4;
            this.label1.Text = "Port Number";
            // 
            // runningLabel
            // 
            this.runningLabel.AutoSize = true;
            this.runningLabel.Location = new System.Drawing.Point(59, 179);
            this.runningLabel.Name = "runningLabel";
            this.runningLabel.Size = new System.Drawing.Size(75, 13);
            this.runningLabel.TabIndex = 5;
            this.runningLabel.Text = "Running: false";
            this.runningLabel.TextAlign = System.Drawing.ContentAlignment.TopCenter;
            // 
            // timer
            // 
            this.timer.Interval = 1000;
            this.timer.Tick += new System.EventHandler(this.timer_Tick);
            // 
            // logView
            // 
            this.logView.Location = new System.Drawing.Point(171, 14);
            this.logView.Multiline = true;
            this.logView.Name = "logView";
            this.logView.ReadOnly = true;
            this.logView.Size = new System.Drawing.Size(291, 179);
            this.logView.TabIndex = 6;
            // 
            // portNumberHttps
            // 
            this.portNumberHttps.Location = new System.Drawing.Point(49, 59);
            this.portNumberHttps.MaxLength = 5;
            this.portNumberHttps.Name = "portNumberHttps";
            this.portNumberHttps.Size = new System.Drawing.Size(100, 20);
            this.portNumberHttps.TabIndex = 7;
            this.portNumberHttps.Text = "8082";
            this.portNumberHttps.WordWrap = false;
            // 
            // label2
            // 
            this.label2.AutoSize = true;
            this.label2.Location = new System.Drawing.Point(17, 36);
            this.label2.Name = "label2";
            this.label2.Size = new System.Drawing.Size(27, 13);
            this.label2.TabIndex = 8;
            this.label2.Text = "Http";
            // 
            // label3
            // 
            this.label3.AutoSize = true;
            this.label3.Location = new System.Drawing.Point(12, 62);
            this.label3.Name = "label3";
            this.label3.Size = new System.Drawing.Size(32, 13);
            this.label3.TabIndex = 9;
            this.label3.Text = "Https";
            // 
            // Form1
            // 
            this.AutoScaleDimensions = new System.Drawing.SizeF(6F, 13F);
            this.AutoScaleMode = System.Windows.Forms.AutoScaleMode.Font;
            this.ClientSize = new System.Drawing.Size(474, 205);
            this.Controls.Add(this.label3);
            this.Controls.Add(this.label2);
            this.Controls.Add(this.portNumberHttps);
            this.Controls.Add(this.logView);
            this.Controls.Add(this.runningLabel);
            this.Controls.Add(this.label1);
            this.Controls.Add(this.portNumberHttp);
            this.Controls.Add(this.init);
            this.Controls.Add(this.stop);
            this.Controls.Add(this.start);
            this.FormBorderStyle = System.Windows.Forms.FormBorderStyle.FixedDialog;
            this.Name = "Form1";
            this.Text = "Form1";
            this.ResumeLayout(false);
            this.PerformLayout();

        }

        #endregion

        private System.Windows.Forms.Button start;
        private System.Windows.Forms.Button stop;
        private System.Windows.Forms.Button init;
        private System.Windows.Forms.TextBox portNumberHttp;
        private System.Windows.Forms.Label label1;
        private System.Windows.Forms.Label runningLabel;
        private System.Windows.Forms.Timer timer;
        private System.Windows.Forms.TextBox logView;
        private System.Windows.Forms.TextBox portNumberHttps;
        private System.Windows.Forms.Label label2;
        private System.Windows.Forms.Label label3;
    }
}

