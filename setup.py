import setuptools


with open("README.md") as fp:
    long_description = fp.read()


setuptools.setup(
    name="sqs_fargate_poller",
    version="0.4.0",

    description="An empty CDK Python app",
    long_description=long_description,
    long_description_content_type="text/markdown",

    author="Marek Kuczynski, marekq@",

    package_dir={"": "sqs_fargate_poller"},
    packages=setuptools.find_packages(where="sqs_fargate_poller"),

    install_requires=[
        "aws-cdk.core==1.31.0",
    ],

    python_requires=">=3.8",

    classifiers=[
        "Development Status :: 4 - Beta",

        "Intended Audience :: Developers",

        "License :: OSI Approved :: Apache Software License",

        "Programming Language :: Python :: 3.8",

        "Topic :: Software Development :: Code Generators",
        "Topic :: Utilities",

        "Typing :: Typed",
    ],
)
