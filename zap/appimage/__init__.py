import os
import platform
import stat

from zap.constants import COMMAND_WRAPPER
from zap.utils import download_file

MACHINE = platform.machine()


class AppImageCore:
    def __init__(self, json_file):
        self._json = json_file
        self._owner = json_file.get('owner')
        self._source = json_file.get('source')
        self._type = self._source.get('type')
        self._url = self._source.get('url')

    def get_release_by_id(self, uid):
        return self._json.get(str(uid))

    def get_latest_stable_release(self):
        result = None
        for i in self._json:
            if not i.isdigit():
                continue
            if self._json[i].get('prerelease'):
                continue
            result = self._json[i]
            break
        return result

    def get_latest_prerelease(self):
        result = None
        for i in self._json:
            if not i.isdigit():
                continue
            result = self._json[i]
        return result

    @staticmethod
    def get_release_assets(data, show_all=False):
        assets = data.get('assets')
        assets_data = dict()
        for asset in assets:
            if platform.machine() in assets[asset].get('name'):
                assets_data[asset] = assets[asset]
        if len(assets_data) == 0 or show_all:
            return assets
        else:
            return assets_data

    def install(self, data, directory, name=False):
        print("Installing {}".format(data.get('name')))
        downloaded_file = download_file(
            data.get('download'),
            output_directory=directory, left_description=' ')
        os.chmod(downloaded_file, 0o755)
        print("Downloaded {file} from {author}".format(
            file=data.get('download'),
            author=self._owner
        ))
        if name:
            downloaded_file_absolute_name = name
        else:
            downloaded_file_absolute_name = \
                downloaded_file.split(os.path.sep)[-1].lower().split('-')[0]

        command_wrapper_file_path = \
            os.path.join(directory, 'bin', downloaded_file_absolute_name)

        with open(command_wrapper_file_path, 'w') as fp:
            fp.write(COMMAND_WRAPPER.format(path_to_appimage=downloaded_file))
        os.chmod(command_wrapper_file_path, 0o755)
        return {
            'path': downloaded_file,
            'name': downloaded_file_absolute_name
        }

    def latest_releases(self):
        releases = dict()
        for app in self._json:
            if not app.isdigit():
                continue
            if 'untagged' in self._json[app].get('tag'):
                break
            releases[app] = self._json[app].get('tag')
            if '.' in self._json[app].get('tag') or \
                    self._json[app].get('tag').isdigit():
                break
        return releases
