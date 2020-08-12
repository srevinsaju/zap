#!/usr/bin/env python3
"""
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

-----------------------------
This file is part of Zap AppImage Package Manager
"""

import json
import os
import shlex
import subprocess
import sys
import urllib.parse
import webbrowser
import re

from halo import Halo

from .appimage.generator import AppImageConfigJsonGenerator
from .libappimage.libappimage import LibAppImage, LibAppImageNotFoundError, \
    LibAppImageRuntimeError
from .appimage import AppImageCore
from .config.config import ConfigManager
from .constants import URL_ENDPOINT, YES_RESPONSES, URL_SHOWCASE, \
    COMMAND_WRAPPER, BUG_TRACKER
from .utils import format_colors as fc
import requests


def parse_gh_url(url):
    """Should input something like this{
        "type": "GitHub",
        "url": "AkashaProject/Community"
    },
    {
        "type": "Download",
        "url": "https://github.com/AkashaProject/Community/releases"
    }"""
    root_url = urllib.parse.urlparse(url)
    url_path = root_url.path[1:] if \
        root_url.path.startswith('/') else root_url.path
    data = [
        {
            'type': "GitHub",
            'url': '{}'.format(url_path)
        }
    ]
    ap = AppImageConfigJsonGenerator({'links': data})
    cb_data = ap.get_app_metadata()
    if cb_data:
        print("Valid GitHub URL Found")
        print(fc("{r}WARNING: Installing untested appimages are not "
                 "suggested. These AppImages have not been tested on a LTS "
                 "Ubuntu distribution. Use at your own risk{rst}"))
        return cb_data
    else:
        print("Parsing GitHub URL failed. Probably the link provided is "
              "non-existent or might not have a valid appimage release.")
        print("If you think, this has been an error, consider creating an "
              "issue at {}".format(BUG_TRACKER))
        return False


class Zap:
    def __init__(self, app):
        self.app = app.lower()
        if os.getenv('APPIMAGE'):
            self.add_self_to_path()
        self.cfgmgr = ConfigManager()

    def add_self_to_path(self, force=True):
        xdg_user_local_dir = \
            os.path.join(os.path.expanduser('~'), '.local', 'bin')
        if xdg_user_local_dir not in os.getenv('PATH').split(os.pathsep):
            print("Warning: {} is not on PATH. Consider adding it to PATH "
                  "for a better experience")
            print("Fallback to {}".format(self.cfgmgr['bin']))
            xdg_user_local_dir = self.cfgmgr['bin']

        if not os.path.exists(xdg_user_local_dir):
            os.makedirs(xdg_user_local_dir)

        zap_bin_file = os.path.join(xdg_user_local_dir, 'zap')
        if not os.path.exists(zap_bin_file) or force:
            with open(zap_bin_file, 'w') as w:
                w.write(COMMAND_WRAPPER.format(
                    path_to_appimage=os.getenv('APPIMAGE')))


    @property
    def is_installed(self):
        return os.path.exists(self.app_data_path)

    @property
    def app_data_path(self):
        return \
            os.path.join(self.cfgmgr['database'], '{}.json'.format(self.app))

    def check_app_installed_verbose(self):
        # check if the release is already installed
        if self.is_installed:
            print("{} is already installed and configured. ".format(self.app))
            return True
        return False

    def remove(self):
        """
        Removes the app
        :return:
        :rtype:
        """
        if not self.is_installed:
            # The app is not installed
            print("{} is not yet installed.".format(self.app))
            return
        response = input("Are you sure you want to remove {}? "
                         "(Y/n)".format(self.app))
        if response not in YES_RESPONSES:
            print("Aborted!")
            return

        appdata = self.appdata()

        appimage_path = appdata.get('path')
        bin_path = os.path.join(self.cfgmgr.bin, appdata.get('name'))
        config_path = self.app_data_path
        for path in (appimage_path, bin_path, config_path):
            if os.path.exists(path):
                os.remove(path)
        try:
            libappimage = LibAppImage()
            if libappimage.is_registered_in_system(appimage_path):
                libappimage.unregister_in_system(appimage_path)
        except LibAppImageNotFoundError:
            pass
        except LibAppImageRuntimeError:
            print("Removing desktop integration failed. libappimage.so "
                  "failed with some errors. Consider removing it manually.")

        print("{} successfully uninstalled!")

    @staticmethod
    def _iter_releases_show_tags_stdout_get_choice(releases,
                                                   select_default=False):
        default = "0"
        for i in releases:
            default = i
            print("[{i}] {name}".format(i=i, name=releases[i]))
        print()

        if len(releases) > 1:
            if select_default:
                choice = "0"
            else:
                choice = input('Select: ').strip().lower()
            if choice not in releases:
                print("Invalid selection.")
                sys.exit(1)
        else:
            print("Selecting latest release {}".format(releases[default]))
            choice = default
        return choice

    @staticmethod
    def _iter_releases_show_assets_stdout_get_choice(assets,
                                                     select_default=False):
        default = "0"
        for i, name in enumerate(assets):
            default = i
            print("[{i}] {name}".format(
                i=i, name=assets.get(name).get('name')))

        if len(assets) > 1:
            if select_default:
                tag_release_asset_choice = "0"
            else:
                tag_release_asset_choice = input('Select: ').strip().lower()
            if not tag_release_asset_choice.isdigit():
                print("Invalid input.")
                sys.exit(1)
            if int(tag_release_asset_choice) not in dict(enumerate(assets)):
                print("Invalid selection.")
                sys.exit(1)
        else:
            tag_release_asset_choice = default
        return tag_release_asset_choice

    def install(self,
                select_default=False, force_refresh=False,
                executable=False, tag_name=None, download_file_in_tag=None,
                always_proceed=False, cb_data=None):
        """
        Installs the app and configures it
        :return:
        :rtype:
        """
        is_app_installed = self.check_app_installed_verbose()
        if is_app_installed and not force_refresh:
            # already installed, so skip!
            return

        if force_refresh:
            print("Force updating appimage...")

        if cb_data is None:
            print("Fetching information for {}".format(self.app))
            r = requests.get(URL_ENDPOINT.format(self.app))
            if not r.status_code == 200:
                # the app does not exist or the name provided is incorrect
                print("Sorry. The app does not exist on our database.")
                return False

            result_core_api = r.json()
        elif isinstance(cb_data, dict):
            result_core_api = cb_data
        else:
            raise ValueError("Invalid data was provided as cb_data")

        core = AppImageCore(result_core_api)
        releases = core.latest_releases()

        if len(releases) < 1:
            print("The app does not provide any releases.")
            return False

        if tag_name and tag_name in releases.values():
            for _tag_id, _tag_name in releases.items():
                if tag_name == _tag_name:
                    tag_choice = _tag_id
                    break
            else:
                raise RuntimeError("Tag was found in releases, "
                                   "but could not be extracted")
        else:
            tag_choice = self._iter_releases_show_tags_stdout_get_choice(
                releases, select_default=select_default)

        release = core.get_release_by_id(tag_choice)
        assets = core.get_release_assets(release)

        if download_file_in_tag and download_file_in_tag in assets.keys():
            tag_release_asset_choice = download_file_in_tag
            asset_data = assets[tag_release_asset_choice]
        else:
            tag_release_asset_choice = \
                self._iter_releases_show_assets_stdout_get_choice(
                    assets=assets, select_default=select_default
                )
            asset_data = \
                assets[dict(enumerate(assets))[int(tag_release_asset_choice)]]

        print("Download size: {}".format(asset_data.get('size')))
        if not always_proceed:
            y = input("Proceed? (Y/n) ")
            if y.lower() not in YES_RESPONSES:
                print("Terminating on user request")
                return False

        cb_data = core.install(
            asset_data,
            self.cfgmgr.storageDirectory,
            name=self.app if not executable else executable)

        print("Configuring...")
        cb_data['uid'] = release.get('id')
        # write data
        with open(self.app_data_path, 'w') as w:
            json.dump(cb_data, w)

        print("Integrating with desktop...")
        try:
            libappimage = LibAppImage()
            if libappimage.is_registered_in_system(cb_data.get('path')):
                libappimage.unregister_in_system(cb_data.get('path'))
            libappimage.register_in_system(cb_data.get('path'))
        except LibAppImageNotFoundError:
            print("Warning: libappimage.so was not found on your system. "
                  "For better features, like desktop integration, consider "
                  "installing libappimage or install zap as an AppImage...")
        except LibAppImageRuntimeError:
            print("Warning: libappimage.so was found on the host system but "
                  "failed to execute some functions due to some errors. "
                  "Consider using Zap Appimage instead of a pip module.")

        print("Done!")

    def _check_for_updates_with_appimageupdatetool(self, path_appimageupdate):
        path_to_old_appimage = self.appdata().get('path')
        spinner = Halo('Checking for updates', spinner='dots')
        spinner.start()
        _check_update_command = shlex.split(
            "{au} --check-for-update {app}".format(
                au=path_appimageupdate,
                app=path_to_old_appimage,
            )
        )
        _check_update_proc = subprocess.Popen(
            _check_update_command,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE
        )
        e_code = _check_update_proc.wait(600)
        if e_code == 0:
            spinner.succeed("Already up-to-date!")
            return
        elif e_code == 1:
            spinner.info("Updates found")
        else:
            spinner.fail("Update information is not embedded within the "
                         "AppImage. ")
            spinner.fail("Consider informing the AppImage author to add a "
                         ".zsync file")
            spinner.fail("Alternatively, pass the --no-appimageupdate option")
        spinner.stop()

    def _update_with_appimageupdatetool(self, path_appimageupdate):
        path_to_old_appimage = self.appdata().get('path')
        spinner = Halo('Checking for updates', spinner='dots')
        spinner.start()
        _check_update_command = shlex.split(
            "{au} --check-for-update {app}".format(
                au=path_appimageupdate,
                app=path_to_old_appimage,
            )
        )

        _check_update_proc = subprocess.Popen(
            _check_update_command,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE
        )
        e_code = _check_update_proc.wait(600)
        if e_code == 0:
            spinner.succeed("Already up-to-date!")
            return
        elif e_code == 1:
            spinner.info("Updates found")
            spinner.start("Updating {}".format(self.app))
            _update_proc = subprocess.Popen(
                shlex.split("{au} --remove-old {app}".format(
                    au=path_appimageupdate,
                    app=path_to_old_appimage,
                )),
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE
            )
            _update_proc_e_code = _update_proc.wait(5000)
            _update_proc_out, _update_proc_err = \
                (x.decode() for x in _update_proc.communicate())
            if _update_proc_e_code == 0:
                # update completed successfully
                spinner.succeed("Update Successful!")
                spinner.start("Setting up new AppImage")
                _file = re.findall(r"Target file: (.*)", _update_proc_out)

                if len(_file) == 1:
                    output_file = _file[0]
                    spinner.info("New file name is {}".format(output_file))
                    _cb_data = self.appdata()
                    _cb_data['path'] = output_file
                    directory = self.cfgmgr.local_storage
                    command_wrapper_file_path = \
                        os.path.join(directory, 'bin', self.app)

                    with open(self.app_data_path, 'w') as w:
                        json.dump(_cb_data, w)

                    with open(command_wrapper_file_path, 'w') as fp:
                        fp.write(COMMAND_WRAPPER.format(
                            path_to_appimage=output_file))
                    spinner.start("Configuring desktop files...")
                    try:
                        libappimage = LibAppImage()
                        if libappimage.is_registered_in_system(
                                path_to_old_appimage):
                            libappimage.unregister_in_system(
                                path_to_old_appimage)
                        libappimage.register_in_system(output_file)
                    except LibAppImageRuntimeError:
                        pass  # TODO: add some more stuff here
                    except LibAppImageNotFoundError:
                        pass  # TODO: add some more stuff here
                    spinner.succeed("Done!")
                else:
                    spinner.stop()
                    print(_file)
                    raise RuntimeError("More than one link found")
            else:
                # Was unsuccessful
                spinner.fail("Update failed! :'(")
                spinner.start("Cleaning up")
                print(_update_proc_out, _update_proc_err)
        else:
            spinner.fail("Update information is not embedded within the "
                         "AppImage. ")
            spinner.fail("Consider informing the AppImage author to add a "
                         ".zsync file")
            spinner.fail("Alternatively, pass the --no-appimageupdate option")

        out, err = (x.decode() for x in _check_update_proc.communicate())
        print(out, err)
        spinner.stop()

    def update(self, use_appimageupdate=True):
        """
        Updates an app using appimageupdate / redownloads the app with new data
        :return:
        :rtype:
        """
        if use_appimageupdate:
            zap_appimageupdate = Zap('appimageupdate')
            zap_appimageupdate.install(select_default=True,
                                       always_proceed=True)
            appimageupdate = zap_appimageupdate.appdata().get('path')
            self._update_with_appimageupdatetool(appimageupdate)
        else:
            raise NotImplementedError()

    def check_for_updates(self, use_appimageupdate=True):
        if use_appimageupdate:
            zap_appimageupdate = Zap('appimageupdate')
            zap_appimageupdate.install(select_default=True,
                                       always_proceed=True)
            appimageupdate = zap_appimageupdate.appdata().get('path')
            self._check_for_updates_with_appimageupdatetool(appimageupdate)
        else:
            raise NotImplementedError

    def show(self):
        """
        Opens a web browser with a tab to the app
        :return:
        :rtype:
        """
        _url = URL_SHOWCASE.format(self.app)
        print(fc("Opening {g}{url}{rst} in browser", url=_url))
        webbrowser.open(_url)

    def appdata(self, stdout=False):
        with open(self.app_data_path, 'r') as r:
            appdata = json.load(r)
        if stdout:
            print(appdata)
        return appdata

    def get_md5(self):
        if not self.is_installed:
            # The app is not installed
            print("{} is not yet installed.".format(self.app))
            return

        lb = LibAppImage()
        path_to_old_appimage = self.appdata().get('path')
        print(path_to_old_appimage)
        print(lb.get_md5(path_to_old_appimage))

    def is_integrated(self):
        if not self.is_installed:
            # The app is not installed
            print("{} is not yet installed.".format(self.app))
            return
        lb = LibAppImage()
        path_to_old_appimage = self.appdata().get('path')
        print(path_to_old_appimage)
        print(lb.is_registered_in_system(path_to_old_appimage))






