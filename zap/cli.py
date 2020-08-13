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
import shutil
import sys

import click
import urllib.parse

from zap.config.config import ConfigManager
from . import __version__
from . import __doc__ as lic
from .zap import Zap, parse_gh_url
from .utils import format_colors as fc


def show_version(ctx, param, value):
    """Prints the version of the utility"""
    if not value or ctx.resilient_parsing:
        return
    click.echo('Zap AppImage utility')
    click.echo('version: {}'.format(__version__))
    ctx.exit()


def show_license(ctx, param, value):
    """Prints the license of the utility"""
    if not value or ctx.resilient_parsing:
        return
    click.echo(lic)
    ctx.exit()


@click.group()
@click.option('--version', is_flag=True, callback=show_version,
              expose_value=False, is_eager=True)
@click.option('--license', '--lic', is_flag=True, callback=show_license,
              expose_value=False, is_eager=True)
def cli():
    """ 🗲 Zap: A command line interface to install appimages"""
    pass


@cli.command()
@click.argument('appname')
@click.option('-d', '--select-default',
              'select_default',  default=False,
              help="Always select first option while installing.")
@click.option('-e', '--executable',
              'executable',  default=False,
              help="Name of the executable, (default: appname)")
@click.option('-f', '--force/--no-force',
              'force_refresh', default=False,
              help="Force install the app without checking.")
def install(appname, **kwargs):
    """Installs an appimage"""
    z = Zap(appname)
    z.install(**kwargs)


@cli.command()
@click.argument('appname')
def remove(appname):
    """Removes an appimage"""
    z = Zap(appname)
    z.remove()


@cli.command()
@click.option('-i', '--interactive/--no-interactive',
              'interactive', default=False,
              help="Interactively edit the configuration")
def config(interactive=False):
    """Shows the config or allows to configure the configuration"""
    cfg = ConfigManager()
    if interactive:
        cfg.setup_config_interactive()
    print(cfg)


@cli.command()
@click.argument('appname')
def appdata(appname):
    """Shows the config of an app"""
    z = Zap(appname)
    z.appdata(stdout=True)


@cli.command()
@click.argument('appname')
@click.option('-a', '--appimageupdate/--no-appimageupdate',
              'use_appimageupdate', default=True,
              help="Use AppImageupdate tool to update apps.")
def update(appname, use_appimageupdate=True):
    """Updates an appimage using appimageupdate tool"""
    z = Zap(appname)
    z.update(use_appimageupdate=use_appimageupdate)


@cli.command()
@click.argument('appname')
@click.option('-a', '--appimageupdate/--no-appimageupdate',
              'use_appimageupdate', default=True,
              help="Use AppImageupdate tool to update apps.")
def check_for_updates(appname, use_appimageupdate=True):
    """Updates an appimage using appimageupdate tool"""
    z = Zap(appname)
    z.check_for_updates(use_appimageupdate=use_appimageupdate)



@cli.command()
def self_update():
    """Update myself"""
    raise NotImplementedError("not yet zapped =)")


@cli.command()
@click.argument('appname')
def show(appname):
    """Get the url to the app and open it in your web browser ($BROWSER)"""
    z = Zap(appname)
    z.show()


@cli.command()
@click.argument('url')
def xdg(url):
    """Parse xdg url"""
    p_url = urllib.parse.urlparse(url)
    query = urllib.parse.parse_qs(p_url.query)
    appname = query.get('app')[0]
    tag = query.get('tag')[0]
    asset_id = query.get('id')[0]
    print(appname, tag, asset_id, type(tag))
    z = Zap(appname)
    if p_url.netloc == 'install':
        print(tag, asset_id)
        z.install(tag_name=tag, download_file_in_tag=asset_id)
    elif p_url.netloc == 'remove':
        z.remove()
    else:
        print("Invalid url")


@cli.command()
@click.argument('appname')
def self_integrate(appname):
    """Add the currently running appimage to PATH, making it accessible
    elsewhere"""
    z = Zap(appname)
    z.add_self_to_path(force=True)


@cli.command()
@click.argument('appname')
def get_md5(appname):
    """Get md5 of an appimage"""
    z = Zap(appname)
    z.get_md5()


@cli.command()
@click.argument('appname')
def is_integrated(appname):
    """Checks if appimage is integrated with the desktop"""
    z = Zap(appname)
    z.is_integrated()


@cli.command()
@click.argument('url')
@click.option('-d', '--select-default',
              'select_default',  default=False,
              help="Always select first option while installing.")
@click.option('-e', '--executable',
              'executable',  default=False,
              help="Name of the executable, (default: last part of url)")
@click.option('-f', '--force/--no-force',
              'force_refresh', default=False,
              help="Force install the app without checking.")
def install_gh(url, executable, **kwargs):
    """Installs an appimage from GitHub repository URL (caution)"""
    # https://stackoverflow.com/q/7160737/
    import re
    regex = re.compile(
        r'^(?:http|ftp)s?://'  # http:// or https://
        r'(?:(?:[A-Z0-9](?:[A-Z0-9-]{0,61}[A-Z0-9])?\.)+'
        r'(?:[A-Z]{2,6}\.?|[A-Z0-9-]{2,}\.?)|'  # domain...
        r'localhost|'  # localhost...
        r'\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})'  # ...or ip
        r'(?::\d+)?'  # optional port
        r'(?:/?|[/?]\S+)$', re.IGNORECASE)

    is_valid_url = re.match(regex, url) is not None
    if not is_valid_url:
        print(fc("{r}Error:{rst} Invalid URL"))
        sys.exit(1)
    cb_data = json.loads(json.dumps(parse_gh_url(url)))
    if executable:
        appname = executable
    else:
        appname = url.split('/')[-1]
    z = Zap(appname)
    z.install(executable=executable, cb_data=cb_data,
              additional_data={'url': url, 'executable': executable},
              **kwargs)


@cli.command()
def disintegrate():
    """Remove zap and optionally remove all the appimages installed with zap"""
    click.confirm('Do you really want to uninstall?', abort=True)
    if click.confirm('Do you want to remove installed AppImages?'):
        cfgmgr = ConfigManager()
        if os.path.exists(cfgmgr['bin']):
            print(fc("{y}Removing bin for appimages{rst}"))
            shutil.rmtree(cfgmgr['bin'], ignore_errors=True)
        if os.path.exists(cfgmgr['storageDirectory']):
            print(fc("{y}Removing storageDirectory for appimages{rst}"))
            shutil.rmtree(cfgmgr['storageDirectory'], ignore_errors=True)
    print(fc("{y}Removing zap binary entrypoint{rst}"))
    for path in os.getenv('PATH').split(os.pathsep):
        zap_bin = os.path.join(path, 'zap')
        if os.path.exists(zap_bin):
            os.remove(zap_bin)
            break
    print(fc("{y}Removing zap AppImage {rst}"))
    dot_zap = os.path.join(os.path.expanduser('~'), '.zap')
    if os.path.exists(dot_zap):
        shutil.rmtree(dot_zap, ignore_errors=True)


if __name__ == "__main__":
    cli()
