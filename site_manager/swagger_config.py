from apispec import APISpec
from apispec.ext.marshmallow import MarshmallowPlugin
from apispec_webframeworks.flask import FlaskPlugin  # type: ignore
from flask import Flask
from flask_swagger_ui import get_swaggerui_blueprint  # type: ignore
from marshmallow import Schema, fields


class ProcessingBodySchema(Schema):
    procedure = fields.Str(description="Procedure: list, status, active, disable or standby",
                           required=True, example="status")
    runService = fields.Str(data_key="run-service", description="Service name to run",
                            required=False, example="serviceA.namespace")
    withDeps = fields.Boolean(data_key="with_deps", description="Print status procedure with dependencies",
                              required=False)


def create_tags(spec: APISpec):
    """
    Tags creation
    :param spec: APISpec object to save tags
    """
    tags = [
        {'name': 'site-manager', 'description': 'Site-manager endpoints'},
    ]

    for tag in tags:
        spec.tag(tag)


def load_docstrings(spec: APISpec, app: Flask):
    """
    Load API descriptions

   :param spec: APISpec object, where descriptions should be loaded
   :param app: Flask app exemplar, where descriptions are taken from
   """
    for fn_name in app.view_functions:
        if fn_name == 'static':
            continue
        view_fn = app.view_functions[fn_name]
        spec.path(view=view_fn)


def get_apispec(app: Flask):
    """ Get api specs

   :param app: flask app
   """
    spec = APISpec(
        title="Site-manager",
        version="1.0.0",
        openapi_version="3.0.3",
        plugins=[FlaskPlugin(), MarshmallowPlugin()],
    )

    jwt_scheme = {"type": "http", "scheme": "bearer", "bearerFormat": "JWT"}

    spec.components.security_scheme("token", jwt_scheme)

    # TODO: add schemas for possible responses
    spec.components.schema("Processing Body", schema=ProcessingBodySchema)

    create_tags(spec)

    load_docstrings(spec, app)

    return spec


SWAGGER_URL = '/docs'
API_URL = '/swagger'

swagger_ui_blueprint = get_swaggerui_blueprint(
    SWAGGER_URL,
    API_URL,
    config={
        'app_name': 'Site-manager'
    }
)
