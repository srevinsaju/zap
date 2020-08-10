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
