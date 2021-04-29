# -*- coding: utf-8 -*-
import matplotlib.pyplot as plt

x =  [5, 7, 9, 11, 13, 15, 17, 19, 21, 23, 25, 27, 29, 31]
x_ticks = [str(v) for v in list(x)]
y_Raft =   [728, 1323, 2150, 3052, 3743,  5589, 6378, 7521, 8389, 9445, 10784, 13451, 15949, 17495]
y_BWRaft = [890, 1123, 1900, 2704, 3443,  5009, 5800, 6971, 7489, 7945,  8851, 9100, 10049, 12095]
y_BFT_BWRaft =  [800, 1423, 2390, 3354, 4043,  5809, 7100, 8271, 9289, 12045, 14584, 15851, 17449, 21095]

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
plt.ylabel("Reach Consensus Time (ms)", font1)  # Y轴标签
# plt.title("A simple plot") #标题
plt.savefig('./consensusTime.jpg')


