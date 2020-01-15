# Syslog CloudWatch Logs bridge

This is a Syslog server that sends all logs received over to [AWS's CloudWatch Logs](https://aws.amazon.com/cloudwatch/details/#log-monitoring).

**Features:**

* Uses AWS's SDK to get credentials from the environment, credentials file or IAM Role.
* TCP and UDP Syslog server on a configurable port (default `5014`).
* Automatic support for syslog messages in RFC3164, RFC6587 or RFC5424 formats.
* Configurable CloudWatch Log Group.
* Creates a new CloudWatch Log Stream on each invocation which is persisted runtime of the server.


## Pre-requirements

* Create an IAM user that can create Log Streams and Logs e.g.

  ```json
  {
    "Version": "2012-10-17",
    "Statement": [
      {
        "Effect": "Allow",
        "Action": [
          "logs:CreateLogStream",
          "logs:PutLogEvents"
      ],
        "Resource": [
          "arn:aws:logs:*:*:*"
      ]
    }
   ]
  }
  ```

* Create new **Log group** on Cloudwatch Logs AWS service.
* On the IoT device, make sure `ntpdate` is installed. If not:

  ```bash
  sudo apt-get install ntpdate
  ```

* (optional) On the IoT device, install `nc` if you want to test forwarding files

  ```bash
  sudo apt-get install netcat
  ```

## Install

Grab a pre-built binary from the [GitHub Releases page](https://github.com/zh/syslog-cloudwatch-bridge/releases).

TODO: release URL here

## Unzip and copy binary

  ```bash
  tar xvzf release----.tar.gz
  cd release-x.y.z
  sudo cp syslog-cloudwatch-bridge /usr/local/bin
  sudo chmod +x /usr/local/bin/syslog-cloudwatch-bridge

  ```

## Config files and start service

From inside the unzipped release directory:

  ```bash
  sudo cp ./scripts/default.cloudwatch /etc/default/cloudwatch
  sudo cp ./scripts/cloudwatch.service /etc/systemd/system/
  sudo service start
  ```

>!!! Set AWS access key, secret key, Log group name and Log stream name in `/etc/default/cloudwatch`

Example: Log Group = *company name*, Log Stream = *IoT device name*

  ```bash
  LOG_GROUP_NAME=company/firm
  LOG_STREAM_NAME=iot-sensor-1
  ```

>!!! Log group should already exists in Cloudwatch Logs AWS service

Check if the service is running on port **5014** (or the post set in the default config).

(optional) Check if the new service accepts and forewards log files (need `nc` installed)

  ```bash
  nc -u 127.0.0.1 5014 <<< "this is CLI log test"

  ```

## Enable service

  ```bash
  sudo service enable cloudwatch
  ```

## Fix rsyslog to forward to Cloudwatch Logs

If you changed the port number in `/etc/default/cloudwatch` change it in `/etc/rsyslog.d/50_cloudwatch.conf` too.


  ```bash
  sudo cp ./scripts/50_rsyslog.conf /etc/rsyslog.d/50_cloudwatch.conf
  sudo service restart rsyslog
  ```

## Check if forwarding is working

  ```bash
  logger -p daemon.info "CLI logging test"
  ```
