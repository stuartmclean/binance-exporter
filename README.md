# binancebot
for doing cool stuff with the binance api
"You know I'm born to lose, and gambling's for fools, but that's the way I like it baby I don't wanna live for ever" - Edward Alan Clarke / Ian Kilmister / Philip John Taylor

## quickstart
- `$ cp .env.example .env`
- fill in the blanks
- `$ make`
- `$ ./binancebot` - for single stats printout
- `$ ./binancebot run` - for regular polling with prometheus and/or influx metrics at frequency defined by POLL_FREQ_SECONDS

## writing to influxdb 1.8 (optional)
Requires `INFLUX_HOST` to be set in .env - otherwise all other influx settings will be ignored.
Only works in `run` mode.

connect to your influxdb via cli `influx`

create user: `CREATE USER username WITH PASSWORD password WITH ALL PRIVILEGES`

create database: `CREATE DATABASE binance`

## Prometheus metrics (optional)
Requires `PROMETHEUS_PORT` to be set in .env
Only works in `run` mode.
Prometheus metrics are generated at `/metrics`.

Example config for prometheus:

```
  - job_name: binancebot
    metrics_path: /metrics
    static_configs:
    - targets: ['10.42.0.19:9000']
```

sample grafana dashboards are provided in [./grafana/].

## Run continously via supervisord
Assuming you are able to run it locally via `$ ./binancebot run`, here is how you would run it on your favourite server in the background by utilizing supervisord [http://supervisord.org/]

The following steps assume you are running FreeBSD, however, the installation should be very similar under any other *nix system.

- install supervisord
    - i.e. under FreeBSD: `pkg install py37-supervisord`
- edit the config file
    - i.e. under FreeBSD: `vim /usr/local/etc/supervisord.conf`
- add the following to the bottom of your `supervisord.conf`:

```
[program:binancebot]
command=/home/YOURUSER/binancebot/binancebot run              ; the program (relative uses PATH, can take args)
directory=/home/YOURUSER/binancebot                ; directory to cwd to before exec (def no cwd)
user=YOURUSER                   ; setuid to this UNIX account to run the program
process_name=%(program_name)s ; process_name expr (default %(program_name)s)
numprocs=1                    ; number of processes copies to start (def 1)
umask=022                     ; umask for process (default None)
autostart=true                ; start at supervisord start (default: true)
autorestart=true              ; retstart at unexpected quit (default: true)
startsecs=10                  ; number of secs prog must stay running (def. 1)
startretries=3                ; max # of serial start failures (default 3)
exitcodes=0,2                 ; 'expected' exit codes for process (default 0,2)
stopsignal=QUIT               ; signal used to kill process (default TERM)
stopwaitsecs=10               ; max num secs to wait b4 SIGKILL (default 10)
redirect_stderr=true          ; redirect proc stderr to stdout (default false)
stdout_logfile=/var/log/binancebot.log        ; stdout log path, NONE for none; default AUTO
stderr_logfile=/var/log/binancebot.err        ; stderr log path, NONE for none; default AUTO
```
- start supervisord
    - i.e. under FreeBSD: in `/etc/rc.conf` add `supervisord_enable="YES"` and run `service supervisord start`
