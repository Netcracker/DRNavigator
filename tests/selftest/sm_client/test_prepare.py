from sm_client.data.structures import *
from sm_client.initialization import init_and_check_config
from sm_client.prepare import make_ordered_services_to_process
from tests.selftest.sm_client.common.test_utils import *


def test_make_ordered_services_to_process():
    init_and_check_config(args_init())

    sm_dict = SMClusterState({
        "site1": {"services": {
            "a": {"after": [], "before": []},
            "b": {"after": [], "before": ["a"]},
            "c": {"after": ["b"], "before": []},
            "d": {"after": ["c"], "before": []},
            "e": {"after": [], "before": ["a"]},
            "f": {"after": [], "before": ["c"]},
            #            "g":{"after":[],"before":["b"]},
        }},
        "site2": {"services": {
            "a": {"after": [], "before": []},
            "b": {"after": [], "before": ["a"]},
            "c": {"after": ["b"], "before": []},
            "d": {"after": ["c"], "before": []},
            "e": {"after": [], "before": ["a"]},
            "f": {"after": [], "before": ["c"]},
            #            "g":{"after":[], "before":["b"]},
        }}})
    sorted_list, code, _ = make_ordered_services_to_process(sm_dict, "site2")
    assert sorted_list == ['b', 'e', 'f', 'a', 'c', 'd'] and code is True

    sm_dict_one_site = SMClusterState({
        "siteN": {"services": {
            "a": {"after": [], "before": []},
            "b": {"after": [], "before": ["a"]},
            "c": {"after": ["b"], "before": []},
            "d": {"after": ["c"], "before": []},
            "e": {"after": [], "before": ["a"]},
        }}})
    sorted_list2, code2, _ = make_ordered_services_to_process(sm_dict_one_site, "siteN")
    assert sorted_list2 == ['b', 'e', 'c', 'a', 'd'] and code2

    sm_dict_absent_deps = SMClusterState({
        "site_with_absent_deps": {"services": {
            "a": {"after": [], "before": []},
            "b": {"after": ["z"], "before": ["a"]},
            "c": {"after": ["a"], "before": ["f"]},
        }}})
    sorted_list3, code3, _ = make_ordered_services_to_process(sm_dict_absent_deps, "site_with_absent_deps")
    assert sorted_list3 == ['b', 'a', 'c'] and code3 is False

    sm_dict_wrong_deps = SMClusterState({
        "site_with_wrong_deps": {"services": {
            "a": {"after": [], "before": []},
            "b": {"after": ["a"], "before": ["a"]},
            "c": {"after": ["a"], "before": ["b"]},
        }}})
    sorted_list4, code4, _ = make_ordered_services_to_process(sm_dict_wrong_deps, "site_with_wrong_deps")
    assert sorted_list4 == [] and code4 is False

    # Test cases for merged sites
    sm_dict_with_dif_deps = SMClusterState({
        "site_1": {"services": {
            "a": {"after": [], "before": []},
            "b": {"after": ["a"], "before": []},
            "c": {"after": [], "before": []},
        },
            "status": True},
        "site_2": {"services": {
            "a": {"after": [], "before": []},
            "b": {"after": [], "before": []},
            "c": {"after": [], "before": ["a"]},
        },
            "status": True}
    })
    sorted_list5, code5, _ = make_ordered_services_to_process(sm_dict_with_dif_deps)
    assert sorted_list5 == ["c", "a", "b"] and code5 is True

    sm_dict_with_cycle_deps = SMClusterState({
        "site_1": {"services": {
            "a": {"after": [], "before": []},
            "b": {"after": ["a"], "before": []}
        },
            "status": True},
        "site_2": {"services": {
            "a": {"after": [], "before": []},
            "b": {"after": [], "before": ["a"]}
        },
            "status": True}
    })
    sorted_list6, code6, _ = make_ordered_services_to_process(sm_dict_with_cycle_deps)
    assert sorted_list6 == [] and code6 is False

    sm_dict_with_cycle_deps_and_not_available_site = SMClusterState({
        "site_1": {"services": {
            "a": {"after": [], "before": []},
            "b": {"after": ["a"], "before": []}
        },
            "status": True},
        "site_2": {"services": {
            "a": {"after": [], "before": []},
            "b": {"after": [], "before": ["a"]}
        },
            "status": False}
    })
    sorted_list7, code7, _ = make_ordered_services_to_process(sm_dict_with_cycle_deps_and_not_available_site)
    assert sorted_list7 == ["a", "b"] and code7 is True
