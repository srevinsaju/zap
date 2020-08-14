import os

from zap.utils import get_executable_path


class Execute:
    def __init__(self, path_to_appimage, use_firejail=False):
        if use_firejail:
            firejail_path = get_executable_path('firejail')
            os.system("{firejail_path} {path_to_appimage}".format(
                firejail_path=firejail_path, path_to_appimage=path_to_appimage
            ))
        else:
            os.system("{path_to_appimage}".format(
                path_to_appimage=path_to_appimage
            ))
