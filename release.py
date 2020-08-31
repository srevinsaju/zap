#!/usr/bin/env python

import os
import sys
from zap import __doc__

print(sys.argv)

if sys.argv[-1].endswith('.py'):
    print("ERR: Invalid argument.")
    sys.exit(-1)

version = sys.argv[-1]
with open(os.path.join('zap', '__init__.py'), 'w') as fp:
    fp.write('"""')
    fp.write(__doc__)
    fp.write('"""')
    fp.write('\n\n')
    fp.write('__version__ = "{}"'.format(version))
    fp.write('\n')
print("Released {}".format(version))
