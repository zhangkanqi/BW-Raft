# -*- coding: utf-8 -*-
import matplotlib.pyplot as plt

x =  [5, 7, 9, 11, 13, 15, 17, 19, 21, 23, 25, 27, 29, 31 ]
x_ticks = [str(v) for v in list(x)]
y_Raft =  [701, 822, 911, 979, 1038,  1129,  1266, 1478, 1456, 1548, 1665, 1655, 1810, 1900]
y_BWRaft = [450, 500, 600, 680, 770, 845, 923, 1004, 1133, 1221, 1344, 1382, 1486, 1515]
y_BFT_BWRaft = [500, 622, 751, 809, 848,  1000,  1046, 1178, 1256, 1318, 1415, 1575, 1590, 1710]

# plt.xlim(20, 260)  # 限定横轴的范围
# plt.ylim(350, 14000)  # 限定纵轴的范围

font1 = {'family': '',
'weight': 'normal',
'size': 17,
}

plt.figure(figsize=(10, 6), dpi=500)

plt.plot(x, y_Raft, marker='*', linewidth=2.7, ms=8, label='Raft',)
plt.plot(x, y_BWRaft, marker='o', linewidth=2.7, ms=8, label='BW-Raft')
plt.plot(x, y_BFT_BWRaft, marker='d', linewidth=2.7, ms=8, label='BFT_BW-Raft')

plt.legend(prop=font1, framealpha=0.2, loc='upper left')  # 让图例生效

plt.xticks(x, x_ticks, rotation=1)

plt.margins(0)
plt.subplots_adjust(bottom=0.2)
plt.xlabel("Cluster Size", font1)  # X轴标签
plt.ylabel("Read Delay (ms)", font1)  # Y轴标签
# plt.title("A simple plot") #标题
plt.savefig('./delay-read.jpg')


