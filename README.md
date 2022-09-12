# SLBR: Suck-Less Bilibili live Recorder

## Why another Bilibili livestream recorder?

Because **all** the other recorders are suck. So I write my own recorder, which, I hope, sucks less.

## Features

If you want a Bilibili recorder featured with:

- Single executable file, just copy and run
- Friendly command-line arguments and an optional configuration file
- Save raw video streams directly, without intentional clipping
- Efficient execution
- Friendly logging to `stdout` or files
- **Just works**

Then you should give SLBR *(suck-less bilibili live recorder)* a try.

## Example usage

### Using config file

```json5
{
  "tasks": [
    {
      // ID of the live room which the task records
      "room_id": 1234,
      "download": {
        // buffer 16MiB data before flushing to disk
        "disk_write_buffer_bytes": 16777216,
        // "." is the default value, you can skip this line
        "save_directory": "."
      },
      "transport": {
        // try ipv4 firstly, then ipv6
        "allowed_network_types": [
          "ipv4",
          "ipv6"
        ]
      }
    }
  ]
}
```

### Using command line arguments

Record live room with `1234` to current working directory:

```shell
./slbr -s 1234 -o .
```

For more usages, run `slbr -h` to get the help menu. Here is a copy (may become outdated):

```
usage: slbr [-h|--help] [-c|--config "<value>"] [-s|--room] [-o|--save-to
            "<value>"] [-b|--disk-write-buffer <integer>]

            Record bilibili live streams

Arguments:

  -h  --help               Print help information
  -c  --config             Specify which configuration file to use. JSON, TOML
                           and YAML are all supported
  -s  --room               Specify which room to record. The ID is included in
                           bilibili live webpage url. Set this to run without
                           config file
  -o  --save-to            Specify the directory where to save records. If not
                           set, process working directory is used
  -b  --disk-write-buffer  Specify disk write buffer size (bytes). The real
                           minimum buffer size is determined by OS. Default:
                           4194304
```

## The project name is too offensive!

You can call it *Simple Lightweight Bilibili live Recorder*. It's all up to you.

## Reference

https://github.com/lovelyyoshino/Bilibili-Live-API

https://github.com/BililiveRecorder/BililiveRecorder

https://github.com/SocialSisterYi/bilibili-API-collect

https://github.com/hr3lxphr6j/bililive-go

https://github.com/zyzsdy/biliroku

## TODO list

- cli override config file
- proper default values for config file
- user manual
