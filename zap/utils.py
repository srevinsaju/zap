import os
import tqdm
import requests

from zap.constants import COLORS


def download_file(
        url, output_directory, left_description,
        filename=None, verbose=False):
    """
    Download file with progressbar
    From https://gist.github.com/ruxi/5d6803c116ec1130d484a4ab8c00c603
    Usage:
        download_file('http://web4host.net/5MB.zip')
    """
    if not filename:
        local_filename = os.path.join(output_directory, url.split('/')[-1])
    else:
        local_filename = os.path.join(output_directory, filename)
    r = requests.get(url, stream=True)
    file_size = int(r.headers['Content-Length'])
    chunk = 1
    chunk_size = 1024
    num_bars = int(file_size / chunk_size)
    if verbose:
        print(dict(file_size=file_size))
        print(dict(num_bars=num_bars))

    with open(local_filename, 'wb') as fp:
        for chunk in tqdm.tqdm(
            r.iter_content(chunk_size=chunk_size), total=num_bars,
            unit='KB', desc=left_description, leave=True  # progressbar stays
        ):
            fp.write(chunk)
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

    return re.match(regex, url) is not None


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
