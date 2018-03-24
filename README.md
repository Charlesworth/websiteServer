# websiteServer

It serves websites.....
- with autogenerated SSL certificates from [Let’s Encrypt](https://letsencrypt.org/)
- with HTTP2 server push for less round trips
- with automatic http to https redirects
- easy static file to path mappings

For an example of a website running on websiteServer visit [ccochrane.com](https://ccochrane.com) and the associated SSL report can be found [here](https://www.ssllabs.com/ssltest/analyze.html?d=ccochrane.com).

Use a mapping.json file to descibe which files to serve at which paths:
```json
{
    "file-paths":[
        {
            "file":"pic.jpg",
            "path":"/pic.jpg"
        },
        {
            "file":"robot.txt",
            "path":"/robot.txt"
        },
    ],
    "push-file-paths":[
        {
            "file":"index.html",
            "path":"/",
            "push-paths":[
                "/pic.jpg"
            ]
        }
    ]
}
```
Then run websiteServer:

    $ ./websiteServer -domain=example.com

Please make sure port 80 and 443 are accessible and DNS for your domain is set up.
A full list of flags can be found via:

    $ ./websiteServer -h

## Building with Docker

An example Dockerfile can be found [here](https://github.com/Charlesworth/charlesworth.github.io).

## Local development
Please have go and go dep installed.

To install dependancies

    $ dep ensure

To build:

    $ go build

Please log any bugs or suggestion as Github issues. Pull requests always welcome.

Special thanks to Let's Encrypt, please consider [donating to them](https://letsencrypt.org/donate/).