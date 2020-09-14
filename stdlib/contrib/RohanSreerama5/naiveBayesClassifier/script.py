#!/usr/bin/env python
# coding: utf-8

# In[70]:


from influxdb_client import InfluxDBClient
import pandas as pd


# In[71]:


mydata = pd.read_csv("zoo_data.csv")


# In[72]:


mydata.head()


# In[73]:


mydata.size


# In[74]:


import datetime
t = pd.date_range(start='1/1/2020', end='05/01/2020', periods=1818)
s = pd.Series(t, name = 'TimeStamp')
mydata.insert(0, 'TimeStamp', s)
mydata = mydata.set_index('TimeStamp')


# In[75]:


mydata.tail()


# In[76]:


token = "UQrVJ9uAC8CZbT79IgegEsvyEb5G-aj7lRfyOeGPeQwKIZTwSVFse93DVdBRsXWRZFpJbRbUC8pxN6Np8diPFQ=="
bucket = "BucketForToday"
org = "hackathonDemoOrg"

from influxdb_client import InfluxDBClient, Point, WriteOptions
from influxdb_client.client.write_api import SYNCHRONOUS

client = InfluxDBClient(url="http://localhost:9999", token=token, org=org, debug=False)
write_client = client.write_api(write_options=SYNCHRONOUS)

write_client.write(bucket, record=mydata, data_frame_measurement_name='zoo-data',
                    data_frame_tag_columns=["Animal_name","airborne"])


# In[ ]:






# %%


# %%
