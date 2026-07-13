from setuptools import setup, find_packages

setup(
    name="limiter-sdk",
    version="1.0.0",
    description="Official Limiter.io Python SDK for rate limiting APIs",
    author="Limiter.io",
    author_email="dev@limiter.io",
    url="https://github.com/neutron420/Limiter.io",
    py_modules=["client", "decorator"],
    install_requires=[
        "requests>=2.28.0",
    ],
    python_requires=">=3.8",
    classifiers=[
        "Programming Language :: Python :: 3",
        "License :: OSI Approved :: MIT License",
        "Operating System :: OS Independent",
    ],
)
