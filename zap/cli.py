import click
import urllib.parse

from zap.config.config import ConfigManager
from . import __version__
from . import __doc__ as lic
from .zap import Zap


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
    """Get md5 of an appimage"""
    z = Zap(appname)
    z.is_integrated()


if __name__ == "__main__":
    cli()
