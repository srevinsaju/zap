import os
from zap import __version__
from setuptools import find_packages
from setuptools import setup


CLASSIFIERS = '''\
Development Status :: 4 - Beta
Intended Audience :: Developers
License :: OSI Approved :: GNU General Public License v3 (GPLv3)
Programming Language :: Python
Topic :: Software Development
Operating System :: POSIX :: Linux
'''


setup(
    name='zap',
    version=__version__,
    packages=find_packages(),
    url='https://github.com/srevinsaju/zap',
    license='MIT',
    author='srevinsaju',
    author_email='srevinsaju@sugarlabs.org',
    description='Zap AppImage Package Manager',
    project_urls={
        'Bug Tracker': 'https://github.com/srevinsaju/zap/issues',
        'Source Code': 'https://github.com/srevinsaju/zap',
    },
    platforms=['Linux'],
    include_package_data=True,
    python_requires='>=3.4',
    entry_points={
        'console_scripts': (
            'zappimage = zap.cli:cli',
            'zap = zap.cli:cli',
        )
    },
    classifiers=[s for s in CLASSIFIERS.split(os.linesep) if s.strip()],
)
