import os
import ctypes

LD_LIBRARY_PATHS = os.getenv('LD_LIBRARY_PATH', '').split(os.pathsep) + \
                   ['/usr/lib', '/usr/local/lib']


def _encode(args):
    return [x.encode() for x in args]


class LibAppImageNotFoundError(Exception):
    pass


class LibAppImageRuntimeError(RuntimeError):
    pass


class LibAppImage:
    def __init__(self):
        for path in LD_LIBRARY_PATHS:
            lib_path = os.path.join(path, 'libappimage.so')
            if os.path.exists(lib_path):
                break
        else:
            raise LibAppImageNotFoundError(
                "libappimage.so was not found in any "
                "of the directories: {}".format(LD_LIBRARY_PATHS))
        self._lib = lib_path
        self._core = ctypes.CDLL(os.path.abspath(self._lib))

    def create_thumbnail(self, *args):
        return self._core.appimage_create_thumbnail(*_encode(args))

    def get_md5(self, *args):
        return self._core.appimage_get_md5(*_encode(args))

    def get_payload_offset(self, *args):
        return self._core.appimage_get_payload_offset(*_encode(args))

    def is_terminal_app(self, *args):
        return self._core.appimage_is_terminal_app(*_encode(args))

    def registered_desktop_file(self, *args):
        return self._core.appimage_registered_desktop_file_path(*_encode(args))

    def register_in_system(self, *args):
        return self._core.appimage_register_in_system(*_encode(args))

    def is_registered_in_system(self, *args):
        return self._core.appimage_is_registered_in_system(*_encode(args))

    def list_files(self, *args):
        return self._core.appimage_list_files(*_encode(args))

    def unregister_in_system(self, *args):
        return self._core.appimage_unregister_in_system(*_encode(args))


