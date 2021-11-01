#!/usr/bin/env python
# -*- coding=utf-8 -*-

import os
import json
import pickle

try:
    import cPickle as pickle
except:
    import pickle


def get_builder(filename):
    with open(filename, "rb") as fb:
        builder = pickle.load(fb)
    return builder


def get_hplist(builder, object=False):
    """Get device information from the builder file.

    Args:
        builder: A builder dict.
        object: Boolean.

    Returns:
        A list.
        [
            {ip: xx,
                port: xx,
                devices: [ # object only
                    xx, ...
                ]
            }
        ]
    """
    result = []
    tmp_dict = {}
    for dev in builder["devs"]:
        if not dev:
            continue
        if int(dev["weight"]) == 0:
            continue

        if dev["ip"] not in tmp_dict:
            tmp_dict[dev["ip"]] = {}
        tmp_dict[dev["ip"]]["port"] = str(dev["port"])

        if object:
            if "devices" not in tmp_dict[dev["ip"]]:
                tmp_dict[dev["ip"]]["devices"] = []
            tmp_dict[dev["ip"]]["devices"].append(dev["device"])

    for ip in tmp_dict.keys():
        if object:
            result.append(
                {
                    "host": ip,
                    "port": tmp_dict[ip]["port"],
                    "devices": tmp_dict[ip]["devices"],
                }
            )
        else:
            result.append({"host": ip, "port": tmp_dict[ip]["port"]})
    return result


builder_info = {}
swift_path = "/etc/swift/"

builder_info["account"] = get_hplist(
    get_builder(os.path.join(swift_path, "account.builder"))
)
builder_info["container"] = get_hplist(
    get_builder(os.path.join(swift_path, "container.builder"))
)
builder_info["object"] = get_hplist(
    get_builder(os.path.join(swift_path, "object.builder")), object=True
)

with open("/etc/swift_exporter.json", "w") as fw:
    fw.write(json.dumps(builder_info))
