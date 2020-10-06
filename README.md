<h1 align="center">
	<img src="./assets/logo.svg" alt="Zap Logo" height=200 width=200 align="middle">
	:zap: Zap 
</h1>

<h3 align="center">
	The delightful package manager for AppImages :package:
</h3>
<div align="center">

[![forthebadge made-with-python](http://ForTheBadge.com/images/badges/made-with-python.svg)](https://www.python.org/)<br/><br/>

![Continuous](https://github.com/srevinsaju/zap/workflows/Continuous/badge.svg) ![GitHub commit activity](https://img.shields.io/github/commit-activity/m/srevinsaju/zap) ![GitHub](https://img.shields.io/github/license/srevinsaju/zap) ![GitHub tag (latest SemVer)](https://img.shields.io/github/v/tag/srevinsaju/zap?sort=semver) [![irc](https://img.shields.io/badge/IRC-%23AppImage%20on%20freenode-blue.svg)](https://webchat.freenode.net/?channels=AppImage) ![Discord](https://img.shields.io/discord/743158702344896574)

[![Mentioned in Awesome AppImage](https://awesome.re/mentioned-badge-flat.svg)](https://github.com/AppImage/awesome-appimage)

[![GitHub followers](https://img.shields.io/github/followers/srevinsaju?label=Follow%20me&style=social)](https://github.com/srevinsaju) [![GitHub stars](https://img.shields.io/github/stars/srevinsaju/zap?style=social)](https://github.com/srevinsaju/zap/stargazers)
</div>



The status of development is still in Beta. Please help me fix bugs by opening an issue.

## Getting Started

### Prerequisites

* A Unix-like operating system: macOS, Linux, BSD. On Windows: `WSL2` is preferred, but `cygwin` or `msys` also mostly work.
* `curl` or `wget` should be installed

### Basic Installation

Zap Package Manager can be installed by running one of the following commands, with `curl` or `wget` or you can manually do the hardwork of download `Zap` AppImage and integrating with your system. 

##### via `curl`

```bash
sh -c "$(curl -fsSL https://git.io/zapinit)"
```

##### via `wget`

```bash
sh -c "$(wget -O- https://git.io/zapinit)"
```

Just copy paste these into the terminal, and you are all ready.

##### Manual Inspection

It's a good idea to inspect the install script from projects you don't yet know. You can do that by downloading the install script first, looking through it so everything looks normal, then running it:

```bash
curl -Lo install_zap.sh https://raw.githubusercontent.com/srevinsaju/zap/master/tools/install.sh
# check the script file for any ambiguity
# for example: cat install_zap.sh ?
sh install_zap.sh
```

<br>

## Using Zap

Zap package manager is extremely easy to use. You can certainly be a pro at using `zap` if you already know how to use `apt` or `dnf`. Cool!

To check if `zap` got installed correctly, just type `zap` on the command line!

```
❯ zap
Usage: zap [OPTIONS] COMMAND [ARGS]...

  🗲 Zap: A command line interface to install appimages

Options:
  --version
  --license, --lic
  --help            Show this message and exit.

Commands:
  appdata            Shows the config of an app
  check-for-updates  Updates an appimage using appimageupdate tool
  config             Shows the config or allows to configure the...
  disintegrate       Remove zap and optionally remove all the appimages...
  get-md5            Get md5 of an appimage
  install            Installs an appimage
  install-gh         Installs an appimage from GitHub repository URL...
  integrate          Checks if appimage is integrated with the desktop
  is-integrated      Checks if appimage is integrated with the desktop
  remove             Removes an appimage
  self-integrate     Add the currently running appimage to PATH, making it...
  self-update        Update myself
  show               Get the url to the app and open it in your web browser...
  update             Updates an appimage using appimageupdate tool
  x                  Execute a Zap installed app (optionally with
                     sandboxing...
  xdg                Parse xdg url
```

Wow, looks cool right?

<br>

### Installing an AppImage

Lets try one of the most interesting, AppImage, [Subsurface](https://subsurface-divelog.org/), for which there is a lot of historical significance, in the beginning of the AppImage concept

```bash
zap install subsurface
```

And you are done!

#### Desktop Integration

Powered by [libappimage](https://github.com/AppImage/libappimage), `zap` supports desktop integration. That means, you don't have to worry about creating desktop files for your downloaded AppImages; everything is so automated.

#### Where does AppImage get downloaded?

By default, `zap` installs the AppImage in `~/.local/share/zap` , which is provided by your `$XDG_DATA_DIR` , you can find all your downloaded copies of the AppImages there

#### Seamless updates

Well, `zap` is an all in one tool which brings the best of all the technologies implemented at the AppImage Org. Using [AppImageUpdate](https://github.com/AppImage/AppImageUpdate), supported AppImages can be updated with `delta` updates using `zsync`. That means you no longer would have to download the entire appimage again for every update, only the changed parts. Well thats awesome. 

By default, `zap` uses `appimageupdate` to update apps. You can manually force `zap` to not use `appimageupdate` by providing `--no-appimageupdate`

```
❯ zap update --help
Usage: zap update [OPTIONS] APPNAME

  Updates an appimage using appimageupdate tool

Options:
  -a, --appimageupdate / --no-appimageupdate
                                  Use AppImageupdate tool to update apps.
  --help                          Show this message and exit.

```

<br>



### Not interested in the command line business?

Well, the new AppImage Catalog (`appimage.github.io - v2`) which is under development has a `Click-2-Install` feature. Just click on the Lightning bolt icon on the Apps release you would like to install.

![image-20200813155429978](/home/srevinsaju/repo/zap/assets/xdg.png)

This feature is still under testing, to enable this feature, please add the following file to this directory

```config
~/.local/share/applications/zap-protocol.desktop
------------------------------------------------
[Desktop Entry]
Name=Zap
Exec=$HOME/zap/zap-x86_64.AppImage xdg %u
Icon=zap
Type=Application
Terminal=true
MimeType=x-scheme-handler/zap;
```

<br>



## Support

All Pull Requests are welcome.

If you are a non-coder or was inspired by this small project, I would be glad if you would :star2: this repository, and spread the word with your friends and foes :smile:



## Credits

This project has been possible with the help and support provided by the AppImage community. Thanks to the detailed responses I received from mentors at AppImage's freenode channel.

Many parts of this documentation have been adapted (~~plagiarized~~) from [The OhMyZsh Project](https://github.com/ohmyzsh/ohmyzsh), you may find the documentation structure is almost similar. And also the installation script.



## License

```
MIT License

Copyright (c) 2020 Srevin Saju

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

## Copyright
(C) Srevin Saju 2020








