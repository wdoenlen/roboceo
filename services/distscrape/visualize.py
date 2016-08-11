'''
Give the output of the distance scraper to this
program to create a heat map visualization.

It works by doing bilinear interpolation between
the sampled distances.

For now it assumes coordinates are in the range
[0, 1). When I finish latitude & longitude output
this will have to be updated.
'''

import sys

from scipy import interpolate
import matplotlib.pyplot as plt
import numpy as np

infile = sys.argv[1]

x = np.array([])
y = np.array([])
z = np.array([])

for line in open(infile):
	row = [float(c) for c in line.split(' ')]
	x = np.append(x, row[0])
	y = np.append(y, row[1])
	z = np.append(z, row[2])

x_grid, y_grid = np.mgrid[0:1:100j, 0:1:200j]

z_grid = interpolate.griddata((x, y), z, (x_grid, y_grid), method='linear')

plt.imshow(z_grid.T, extent=(0,1,0,1), origin='lower')
plt.show()
