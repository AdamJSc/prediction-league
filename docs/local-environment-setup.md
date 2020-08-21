[Back to README](../README.md)

This project can be run either:
* **Fully-Dockerised** - runs exclusively within containers, or
* **Partly-Dockerised** - dependencies and build tools are in containers, but the service runs using npm/nvm and Golang
on the host machine.

As a bare minimum, this project runs a MySQL container as a persistent data store.

The builds comprise Sass for CSS and Babel via Webpack for JavaScript (including Vue).

But "how did we get here?" I ~~hear~~ presume you ask...

## A ~~comedy of errors~~ note about the local environment... 🎭

The idea was originally to use a single process (Webpack) to generate both JavaScript and CSS, but
I couldn't get the Webpack process to produce anything other than standard JS output when building Sass(?)

JS syntax generally isn't valid CSS syntax (fun fact).

According to the obligatory Google/Stack Overflow delve, absolutely no one else in the history of development appeared
to have ever experienced anything like this issue before, so I reverted to running separate Sass and Babel CLIs via npm
commands instead (the natural workaround).

This worked fine... until... I added Vue, of course. (Why would it continue to work fine? I mean, just why?)

At this point, Babel decided it was going to go maverick and not properly transpile a "require" function or something 👍

So, unable to pinpoint exactly what I was doing wrong, and with several hours lost, I opted to reintroduce Webpack
purely for compiling Vue.

Hence why we now have a weird Sass CLI/Webpack mashup combo.

Because it sometimes does what it's supposed to.

I prefer Backend...

Anyway, both build processes will watch for changes to CSS/JS/Vue files - however, changing Go files or HTML
files/templates will require the service itself to be restarted each time.

It is possible to implement an additional Go-based watcher which will recompile the binary each time changes are detected
though. Maybe one for the backlog...
