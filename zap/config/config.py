import json
import os
import appdirs

from zap.constants import ZAP_PATH_RC_PATCH


class ConfigManager:
    def __init__(self):
        self._config_dir = appdirs.user_config_dir('zap')
        self._cfgpath = os.path.join(self._config_dir, 'config.json')
        self._config = {
            'version': '0.1',
            'name': 'ZapConfiguration',
            'storageDirectory': appdirs.user_data_dir('zap'),
            'bin': os.path.join(appdirs.user_data_dir('zap'), 'bin')
        }

        self.make_data_directory()

        if not os.path.exists(self._config_dir):
            self.make_config_directory()
        if not os.path.exists(self._cfgpath):
            self.write_file()

        self.read_file()

    def __getattr__(self, item):
        return self._config.get(item)

    def __getitem__(self, item):
        return self._config.get(item)

    def __setitem__(self, key, value):
        self._config[key] = value
        self.write_file()

    def __repr__(self):
        return 'ZapConfig({})'.format(json.dumps(self._config, indent=4))

    @property
    def local_storage(self):
        return self._config['storageDirectory']

    @property
    def shell_wrappers_dir(self):
        return self._config.get('bin')

    def add_current_directory_to_path(self):
        if os.getenv('ZAP_PATH'):
            # already configured!
            return
        path_to_current_shell = os.getenv('SHELL')
        if path_to_current_shell:
            current_shell = path_to_current_shell.split(os.path.sep)[-1]
            print("Detected {} shell".format(current_shell))
            path_to_shell_rc = os.path.join(os.path.expanduser('~'),
                                            '.{}rc'.format(current_shell))
            if os.path.exists(path_to_shell_rc):
                with open(path_to_shell_rc, 'r') as fp:
                    shrc_file = fp.read()
                if 'ZAP_PATH' in shrc_file:
                    print("Shell is already configured to use zap! ")
                    return
                print("Patching .{}rc file to include AppImage configuration"
                      .format(current_shell))
                with open(path_to_shell_rc, 'a') as fp:
                    fp.write(ZAP_PATH_RC_PATCH.format(path_to_bin=self['bin']))
            else:
                print("Could not find a .bashrc or .zshrc file on home "
                      "directory. Please add the following code to your "
                      "shell configuration to make appimages run anywhere "
                      "on the command line")
                print("Code follows below: \n\n")
                print(ZAP_PATH_RC_PATCH.format(path_to_bin=self['bin']))
        else:
            print("Unable to detect current shell. looks like your $SHELL "
                  "variable is unset. Please set the $SHELL variable or add "
                  "the following code manually to the shell configuration "
                  "file")
            print(ZAP_PATH_RC_PATCH.format(path_to_bin=self['bin']))

    def make_data_directory(self):
        data_directory = self._config.get('storageDirectory')
        bin_directory = os.path.join(data_directory, 'bin')
        db_directory = os.path.join(data_directory, 'db')
        if not os.path.exists(data_directory):
            os.makedirs(data_directory)
        if not os.path.exists(bin_directory):
            os.makedirs(bin_directory)
            self['bin'] = bin_directory
        if not os.path.exists(db_directory):
            os.makedirs(db_directory)
            self['database'] = db_directory
        self.add_current_directory_to_path()

    def make_config_directory(self):
        os.makedirs(self._config_dir)
        self.write_file()

    def write_file(self):
        with open(self._cfgpath, 'w') as w:
            json.dump(self._config, w, indent=4)

    def read_file(self):
        with open(self._cfgpath, 'r') as r:
            self._config.update(json.load(r))

    def setup_config_interactive(self):
        print("Zap AppImage Package Manager configuration wizard")
        print()
        path_to_save_appimages = input("Please specify absolute path to "
                                       "store appimages: ")
        path_to_save_appimages = os.path.abspath(path_to_save_appimages)
        if not os.path.exists(path_to_save_appimages):
            print("Setting up {}".format(path_to_save_appimages))
            os.makedirs(path_to_save_appimages)
        elif os.path.exists(path_to_save_appimages) and \
                os.path.isfile(path_to_save_appimages):
            raise NotADirectoryError("The path you have provided is not a "
                                     "directory. Please provide a valid "
                                     "directory")
        self._config['storageDirectory'] = path_to_save_appimages
        self.write_file()
        print("Done!")
