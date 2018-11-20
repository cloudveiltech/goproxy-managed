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
            this.portNumber = new System.Windows.Forms.TextBox();
            this.label1 = new System.Windows.Forms.Label();
            this.runningLabel = new System.Windows.Forms.Label();
            this.timer = new System.Windows.Forms.Timer(this.components);
            this.logView = new System.Windows.Forms.TextBox();
            this.SuspendLayout();
            // 
            // start
            // 
            this.start.Location = new System.Drawing.Point(59, 98);
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
            this.stop.Location = new System.Drawing.Point(59, 127);
            this.stop.Name = "stop";
            this.stop.Size = new System.Drawing.Size(75, 23);
            this.stop.TabIndex = 1;
            this.stop.Text = "Stop";
            this.stop.UseVisualStyleBackColor = true;
            this.stop.Click += new System.EventHandler(this.stop_Click);
            // 
            // init
            // 
            this.init.Location = new System.Drawing.Point(59, 69);
            this.init.Name = "init";
            this.init.Size = new System.Drawing.Size(75, 23);
            this.init.TabIndex = 2;
            this.init.Text = "Init";
            this.init.UseVisualStyleBackColor = true;
            this.init.Click += new System.EventHandler(this.init_Click);
            // 
            // portNumber
            // 
            this.portNumber.Location = new System.Drawing.Point(49, 33);
            this.portNumber.MaxLength = 5;
            this.portNumber.Name = "portNumber";
            this.portNumber.Size = new System.Drawing.Size(100, 20);
            this.portNumber.TabIndex = 3;
            this.portNumber.Text = "8081";
            this.portNumber.WordWrap = false;
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
            this.runningLabel.Location = new System.Drawing.Point(59, 171);
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
            this.logView.Size = new System.Drawing.Size(291, 167);
            this.logView.TabIndex = 6;
            // 
            // Form1
            // 
            this.AutoScaleDimensions = new System.Drawing.SizeF(6F, 13F);
            this.AutoScaleMode = System.Windows.Forms.AutoScaleMode.Font;
            this.ClientSize = new System.Drawing.Size(474, 193);
            this.Controls.Add(this.logView);
            this.Controls.Add(this.runningLabel);
            this.Controls.Add(this.label1);
            this.Controls.Add(this.portNumber);
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
        private System.Windows.Forms.TextBox portNumber;
        private System.Windows.Forms.Label label1;
        private System.Windows.Forms.Label runningLabel;
        private System.Windows.Forms.Timer timer;
        private System.Windows.Forms.TextBox logView;
    }
}

