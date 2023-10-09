# autobrr-no-rars
This project aims to stop autobrr from downloading torrents with `.rar` files. I've set up two options for doing this: a modified autobrr image that can be hot-swapped with the standard version, and a sidecar container with an API that runs independently of your autobrr instance. Both of these methods will allow you to either limit how many `.rar` files you want to allow in a single torrent, or outright block any torrent with a `.rar` file in it.

## Implementation
There are two folders here: `api/` and `standalone/`. Each one contains a Docker container that provides a method for checking if a torrent contains `.rar` files. You can then configure autobrr to run this check when a torrent is grabbed by a filter, then forward the torrent onwards to your filter's action if it passes this check otherwise stop the torrent there. 
### Lazy way
The `standalone/` folder has a tweaked version of autobrr that has some extra stuff baked in to perform the `.rar` check. **This is super easy to get started**, as all you have to do is swap the base image for your existing autobrr container and add the check to your filter. The downside is that you can't change what version of autobrr you want to use since it's built in to the image. This is normally tolerable but if you want to have control over versioning (usually for stability concerns), then the other option would be better for you.

The tutorial for setting this method up is outlined in `standalone/README.md`.
### Better way
The `api/` folder has a container that exposes an API which can do this check for `.rar` files. It takes a bit more setup since you have to configure the networking for this container but the benefit is that you can use whatever version of autobrr you like. 

Similar to the easy method, the walkthrough for how to set this up is outlined in `api/README.md`.

## Justification and explanation
I got tired of autobrr downloading stuff that has `.rar` files. You have to either extract the `.rar`, which wastes space if you normally use hardlinks to your media folder from your torrent client's download folder (like with Sonarr/Radarr), or you could use something like [rar2fs](https://github.com/hasse69/rar2fs) and try to automate that - both of which are equally unappealing. In autobrr, you can theoretically use tags from whatever indexer you're using to limit what you grab, but most IRC announce channels don't send much data other than the name of the torrent and how to get it. I asked for help with this and the only response I got was to manually blacklist release groups that drop rar'd releases, which sounds like a tedious game of whack-a-mole that I would prefer not to play. I wanted a solution that was agnostic of as much as possible, meaning that it would work on any indexer, release group, download client, etc. All I care about is that I don't get rar'd releases.

All my homies hate rar'd releases.
