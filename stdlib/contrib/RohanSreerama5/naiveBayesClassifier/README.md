# Naive-Bayes-Classifier-Flux
Implementation of Naive Bayes Classifier in Flux

Awesome! So you want to get started with ML in Flux? Let's dive in! 

Steps: 

Prerequesites: InfluxDB 2.0 (local or Cloud instance), this repo (includes zoo-data.csv, script.py and naiveBayesClassifier.flux) 
InfluxDB: https://github.com/influxdata/influxdb

1) Ensure you have InfluxDB 2.0 set up by either going to localhost:9999 (if you `make` InfluxDB from source) or going to your Cloud instance on AWS, Azure, etc. Keep the instance running throughout this demo.  
2) Choose a binary dataset (fields can only take on 2 unique values) or visit https://archive.ics.uci.edu/ml/datasets.php for some wonderful datasets you can get started with. 

![](images/csvData.png)

3) You will have to make edits to script.py in the following areas: 

- Set `mydata` equal to the file path of your dataset 
- Fill in your values for `token`, `bucket`, and `organization`. (If you are using my sample zoo dataset you can ignore the rest of this step) 
- Ensure that `periods` is set equal to the value you get from running the `data.size` cell. 
- Set `dataframe_measurement_name` to be the name of your dataset. 
- Note which fields you want to use as actual fields you classify on and which you would like to use for Class. Recall we predict `P(Class | field)`.
- Anything you choose to be a class must be listed inside of `data_frame_tag_columns` and everything else defaults to a `field`. Finally, run the script. You've just written your dataset to an InfluxDB bucket. 

![](images/pythonScript.png)

Note: In our demo, we've divided training and test data based on time: 3 days for training and 1 day for testing. 

4) In the UI, go to Data -> yourBucket and if data does not show up, then add a `Custom Time Range` that is the time range of your dataset. Switch on `Raw Data`. 
5) Copy-paste in the Flux script from `naiveBayesClassifier.flux` into the `Script Editor`. At the bottom of the script, ensure you have changed the arguments to `naiveBayes(...)` with the correct information from your dataset. If your are running the `zoo-data`, all you must change is the `bucket` name. 

Note: For ease of use and debugging, use the Flux extension in VSCode. https://marketplace.visualstudio.com/items?itemName=influxdata.flux

6) That's it! Hit `Submit` and watch the predictions show up under the `Probability` column. What is this predicting? This classifier predicts the probability that a given animal is airborne given whether it is aquatic or not. 

If you are using my zoo data, feel free to play around with this by changing the field or Class to any of the other available fields in the dataset. You can also change the time frames of the training and test data based on your dataset.

![](images/overview.png)

![](images/data.png)

Coming very soon:

- Support for multiple fields
- Encapsulation of training code into its own function so that training and testing can be performed independently
- Ability to save training models in InfluxDB bucket
- Blog post and ReadMe outlining demo and inspiration

In the works for the future:

- Implementation for common density functions in order to support non-binary datasets. We hope to work on implementing a Gaussian density function in order to be able to calculate probabilities accurately for data with fields that can take on more values than 0 and 1.
- Ability to save training models in SQL table
- Utilize our algorithm to classify Slack incidents (our original goal).
- Create a GUI that allows users to feed in their datasets easily.

Well, you did it! Your first stab at machine learning in Flux. What did you think? I plan on pushing more updates to our implementation in the future and classify more purpose-driven datasets so stay in the loop! 

Huge thanks to Adam Anthony and Anais Dotis-Georgiou for their invaluable guidance and support during this project. And much love to Team Magic: Mansi Gandhi, Rose Parker, and me. 

Be sure to follow InfluxData for more cool demos!

Visit us at: https://www.influxdata.com/

Find more demos at: https://www.influxdata.com/blog/

