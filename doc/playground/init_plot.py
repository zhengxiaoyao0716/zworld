#!/usr/bin/python
# -*- coding: utf-8 -*-

"""
初始化绘图
@author: zhengxiaoyao0716
"""

import sys
from pylab import mpl
import matplotlib.pyplot as plt
from mpl_toolkits.mplot3d import Axes3D as _


def run(main, *args, **kwargs):
    """执行main函数"""
    mpl.rcParams['font.sans-serif'] = ['SimHei']
    mpl.rcParams['axes.unicode_minus'] = False
    fig = plt.figure(figsize=(16, 9), dpi=70)
    main(fig, *args, **kwargs)
    sys.stdout.flush()
    plt.show()


if __name__ == '__main__':
    run(lambda fig: None)
