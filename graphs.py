# csv format:
# transaction_id, time (nanoseconds)

import pandas as pd
import matplotlib.pyplot as plt

NUM_FILES = 3

DATA_FILE_NAME = 'node1_transactions.csv'

df = pd.read_csv(DATA_FILE_NAME)

print(df)
