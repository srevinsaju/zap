import os
from zap.constants import COLORS
from downloader_cli.download import Download


def download_file(url, output_directory, filename=None):
    if not filename:
        local_filename = os.path.join(output_directory, url.split('/')[-1])
    else:
        local_filename = os.path.join(output_directory, filename)
    dw = Download(url, des=local_filename, overwrite=True, )
    dw.download()
    return local_filename


def format_colors(string, **kwargs):
    return string.format(**kwargs, **COLORS)


def is_valid_url(url):
    import re
    regex = re.compile(
        r'^(?:http|ftp)s?://'  # http:// or https://
        r'(?:(?:[A-Z0-9](?:[A-Z0-9-]{0,61}[A-Z0-9])?\.)+'
        r'(?:[A-Z]{2,6}\.?|[A-Z0-9-]{2,}\.?)|'  # domain...
        r'localhost|'  # localhost...
        r'\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})'  # ...or ip
        r'(?::\d+)?'  # optional port
        r'(?:/?|[/?]\S+)$', re.IGNORECASE)
    file_regex = re.compile(
        r'^(?:file)s?://'
        r'(?:/?|[/?]\S+)$'
    )
    return re.match(regex, url) is not None or \
        re.match(file_regex, url) is not None


def get_executable_path(executable, raise_error=True):
    """
    Returns the absolute path of the executable
    if it is found on the PATH,
    if it is not found raises a FileNotFoundError
    :param executable:
    :return:
    """
    # gets the PATH environment variable
    path = os.getenv('PATH').split(os.pathsep)
    for i in path:
        if os.path.exists(os.path.join(i, executable)):
            return os.path.join(i, executable)
    if raise_error:
        raise FileNotFoundError(
            "Could not find {p} on PATH. "
            "Make sure {p} is added to path and try again".format(
                p=executable))
    else:
        return False
