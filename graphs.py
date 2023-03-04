# csv format:
# transaction_id, time (nanoseconds)

import pandas as pd
import matplotlib.pyplot as plt

NUM_FILES = 3

DATA_FILE_NAME = "node{}_transactions_log.csv"


allDfs = []

for i in range(1, NUM_FILES+1):
    df = pd.read_csv(DATA_FILE_NAME.format(i))
    # print(df['time'])
    allDfs.append(df)

df = pd.concat(allDfs, axis=0)


df['time'] = df['time'].apply(
    lambda time: int(time))
print(df['time'])

result = df.groupby('transaction_id').agg({'time': ['min', 'max']})

print(result)


