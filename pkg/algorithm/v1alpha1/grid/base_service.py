import base64
import logging
import operator
import numpy as np
from api.v1alpha1.grpc.python3 import api_pb2
logger = logging.getLogger(__name__)


class Parameter:

    def __init__(self, name, space_list):
        self.name = name
        self.space_list = space_list
        self.space_list.sort()
        self.length = int(len(self.space_list))

    def __str__(self):
        return "Parameter(name: {}, list: {})".format(self.name, ", ".join(self.space_list))


def num2str(assignments, num):
    assignments.sort(key=operator.attrgetter('key'))
    result = ""
    for i in range(num):
        result += str(assignments[i].key) + ": " + str(assignments[i].value)
        if i < num - 1:
            result += '-'
    return result


class BaseSamplingService(object):

    def __init__(self, request):
        self.space = []
        self.space_size = 1
        for _par in request.parameters:
            new_par = Parameter(_par.name, _par.feasible_space)
            self.space.append(new_par)
            self.space_size *= new_par.length
        self.space_size = int(self.space_size)
        self.space.sort(key=operator.attrgetter('name'))
        self.existing_trials = {}
        self.num_pars = int(len(request.parameters))

        for _trial in request.existing_results:
            self.existing_trials[num2str(_trial.parameter_assignments, self.num_pars)] = _trial.object_value

    def get_assignment(self, request):
        logger.info("-" * 100 + "\n")
        print("-" * 100 + "\n")
        logger.info("New getSuggestions call\n")
        print("New getSuggestions call\n")
        if request.algorithm_name == "grid":
            return self.get_assignment_grid(request)
        elif request.algorithm_name == "random":
            return self.get_assignment_random(request)
        return []

    def grid_index_search(self, index):
        assignments = []
        for i in range(self.num_pars):
            sub_space_size = 1
            for j in range(i + 1, self.num_pars):
                sub_space_size *= self.space[j].length
            index_ = int((index % (sub_space_size * self.space[i].length)) / sub_space_size)
            assignments.append(api_pb2.KeyValue(key=self.space[i].name, value=self.space[i].space_list[index_]))
        assert num2str(assignments, self.num_pars) not in self.existing_trials
        self.existing_trials[num2str(assignments, self.num_pars)] = -1
        return assignments

    def random_index_search(self):
        while True:
            assignments = []
            for i in range(self.num_pars):
                assignments.append(api_pb2.KeyValue(key=self.space[i].name, value=self.space[i].space_list[np.random.randint(self.space[i].length)]))
            if num2str(assignments, self.num_pars) not in self.existing_trials:
                break
        assert num2str(assignments, self.num_pars) not in self.existing_trials
        self.existing_trials[num2str(assignments, self.num_pars)] = -1
        return assignments

    def get_assignment_grid(self, request):
        assignments_set = []
        next_assignment_index = int(len(request.existing_results))
        for _ in range(request.required_sampling):
            assignments = self.grid_index_search(next_assignment_index)
            assignments_set.append(api_pb2.ParameterAssignments(key_values=assignments))
            next_assignment_index += 1
            for assignment in assignments:
                logger.info("Name = {}, Value = {}, ".format(assignment.key, assignment.value))
                print("Name = {}, Value = {}, ".format(assignment.key, assignment.value))
            logger.info("\n")
        return assignments_set

    def get_assignment_random(self, request):
        assignments_set = []
        for _ in range(request.required_sampling):
            assignments = self.random_index_search()
            assignments_set.append(api_pb2.ParameterAssignments(key_values=assignments))
            for assignment in assignments:
                logger.info("Name = {}, Value = {}, ".format(assignment.key, assignment.value))
                print("Name = {}, Value = {}, ".format(assignment.key, assignment.value))
            logger.info("\n")
        return assignments_set

    @staticmethod
    def encode(name):
        """Encode the name. Chocolate will check if the name contains hyphens.
        Thus we need to encode it.
        """
        return base64.b64encode(name.encode('utf-8')).decode('utf-8')
