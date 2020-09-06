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
This file is part of AppImage Catalog Generator
"""

import hashlib
import html
import json
import urllib.request
import urllib.error
import uuid
from colorama import Fore


class AppImageConfigJsonGenerator:
    def __init__(self, app, token=None):
        self.token = token
        self._title = html.escape(app.get('name', ''))
        self._categories = app.get('categories')
        self._description = app.get('description')
        self._authors = app.get('authors', [])
        self._licenses = app.get('license')
        self._links = app.get('links')
        self._icon = app.get('icons')
        self._screenshots = app.get('screenshots')
        self.github_info = self.get_github_info()

    @property
    def title(self):
        """
        Returns the raw title without formatting, i.e refer title_formatted
        :return:
        :rtype:
        """
        return self._title

    @property
    def links(self):
        """
        Returns the links to the appimage
        :return:
        :rtype:
        """
        return self._links

    def is_github(self):
        """
        Checks if the app-image has its source link from github
        :return:
        :rtype:
        """
        if not self.links:
            return False

        if not len(self._links) >= 1:
            return False

        if not self._links[0].get("type", '').lower() == "github":
            return False

        return True

    def get_github_release_from(self, github_release_api):
        request = urllib.request.Request(github_release_api)
        if self.token:
            request.add_header("Authorization", "token {}".format(self.token))
        try:
            request_url = urllib.request.urlopen(request)
        except urllib.error.HTTPError:
            print(
                Fore.RED +
                "[STATIC][{}][GH] Request to {} failed with 404".format(
                    self.title, github_release_api
                ) + Fore.RESET)
            return False
        status = request_url.status
        if status != 200:
            print(
                Fore.RED +
                "[STATIC][{}][GH] Request to {} failed with {}".format(
                    self.title, github_release_api, status
                ) + Fore.RESET)
            return False
        return request_url

    def get_github_api_data(self):
        """
        Gets the data from api.github.com
        :return:
        :rtype:
        """
        github_release_api = \
            "https://api.github.com/repos/{path}/releases".format(
                path=self._links[0].get("url")
            )

        # get the request urllib response instance or bool
        request_url = self.get_github_release_from(github_release_api)
        # check if request succeeded:
        if not request_url:
            return False

        # read the data
        github_api_data = request_url.read().decode()  # noqa:

        # attempt to parse the json data with the hope that the data is json
        try:
            json.loads(github_api_data)
        except json.decoder.JSONDecodeError:
            return False

        # load the data
        json_data = json.loads(github_api_data)
        return json_data

    def get_github_info(self):
        if not self.is_github():
            # pre check if the appimage is from github, if not, exit
            return False

        print('[GH] Parsing information from GitHub'.format(
            self.title
        ))

        # process github specific code
        owner = self._links[0].get("url", '').split('/')[0]

        # get api entry-point
        data = self.get_github_api_data()

        if not data or isinstance(data, bool):
            # the data we received is ill formatted or can't be processed
            # return False, because at this point, to not raise ValueError
            # and not to stash the build
            return False

        releases_api_json = dict()
        for i, release in enumerate(data):
            # iterate through the data
            # and process each data
            tag_name = release.get("tag_name")
            appimages_assets = dict()
            for asset in release.get("assets"):
                download_url = asset.get('browser_download_url')
                if download_url.lower().endswith('.appimage'):
                    # a valid appimage file found in release assets
                    uid = hashlib.sha256((
                            asset.get('name') + ":" + download_url).encode()).hexdigest()
                    appimages_assets[uid] = {
                        'name': asset.get('name'),
                        'download': download_url,
                        'count': asset.get('download_count'),
                        'size': "{0:.2f} MB".format(
                            asset.get('size') / (1000 * 1000))
                    }

            uid_appimage = hashlib.sha256(
                "{}".format(self._links[0].get("url")).encode()
            ).hexdigest()

            author_json = release.get('author')
            author = None
            if author_json is not None:
                author = author_json.get('login')
            releases_api_json[i] = {
                'id': uid_appimage,
                'author': author,
                'prerelease': release.get('prerelease'),
                'releases': release.get('html_url'),
                'assets': appimages_assets,
                'tag': tag_name,
                'published_at': release.get('published_at')
            }
        releases_api_json['owner'] = owner
        releases_api_json['source'] = {
            'type': 'github',
            'url': "https://github.com/{path}".format(
                    path=self._links[0].get("url"))
        }
        return releases_api_json

    def get_app_metadata(self):
        if self.is_github():
            return self.github_info

