#!/usr/bin/env python3
"""Export the frozen Angular client's wire contract as OpenAPI."""
import json
import re
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
source = (ROOT / "frontend/datacat-ui/src/shared/services/datacat-generated-client.ts").read_text()
interfaces = dict(re.findall(r"export interface I(\w+)\s*\{(.*?)\n\}", source, re.S))
enums = dict(re.findall(r"export enum (\w+)\s*\{(.*?)\n\}", source, re.S))


def schema(kind):
    kind = re.sub(r"\s*\|\s*(undefined|null)", "", kind).strip()
    if kind.endswith("[]"):
        return {"type": "array", "items": schema(kind[:-2])}
    primitives = {"string": {"type": "string"}, "number": {"type": "number"},
                  "boolean": {"type": "boolean"}, "Date": {"type": "string", "format": "date-time"}}
    if kind in primitives:
        return primitives[kind]
    name = kind[1:] if kind.startswith("I") and kind[1:] in interfaces else kind
    if name in interfaces or name in enums:
        return {"$ref": f"#/components/schemas/{name}"}
    return {}


schemas = {}
for name, fields in interfaces.items():
    properties = {}
    for prop, optional, kind in re.findall(r"^\s*(\w+)(\?)?:\s*([^;]+);", fields, re.M):
        properties[prop] = schema(kind)
    schemas[name] = {"type": "object", "properties": properties}
for name, values in enums.items():
    schemas[name] = {"type": "string", "enum": re.findall(r'=\s*"([^"]+)"', values)}

required = {
    "AddDashboardRequest": ["name"],
    "AddPanelRequest": ["title", "type", "rawQuery", "dataSourceId", "layout", "dashboardId"],
    "AddAlertRequest": ["template", "rawQuery", "dataSourceId", "notificationChannelGroupName", "repeatInterval"],
    "AddDataSourceRequest": ["uniqueName", "dataSourceType", "connectionString", "purpose"],
    "AddPluginRequest": ["name", "version", "author", "endpoint"],
    "AddVariableRequest": ["placeholder", "value"],
    "SecretModel": ["key", "value"],
}
for name, fields in required.items():
    if name in schemas:
        schemas[name]["required"] = fields

paths = {}
method_pattern = re.compile(r"^    (\w+)\(([^\n]*)\): Observable<([^\n]+)> \{(.*?)(?=^    (?:protected |\w+\())", re.S | re.M)
for operation, args, output, body in method_pattern.findall(source):
    route = re.search(r'this.baseUrl \+ "([^"?]+)', body)
    method = re.search(r'this.http.request\("(\w+)"', body)
    if not route or not method:
        continue
    path, verb = route.group(1), method.group(1)
    argument_types = dict(re.findall(r"(\w+): ([^,]+)", args))
    parameters = [{"name": name, "in": "path", "required": True, "schema": schema(argument_types.get(name, "string"))}
                  for name in re.findall(r"\{(\w+)\}", path)]
    for name in re.findall(r'url_ \+= "(\w+)=', body):
        kind = next((value for key, value in argument_types.items() if key.lower() == name.lower()), "string")
        parameters.append({"name": name, "in": "query", "required": False, "schema": schema(kind)})
    global_routes = ("/api/v1/user/", "/api/v1/namespace", "/api/v1/plugin/", "/api/v1/data-source-type/",
                     "/api/v1/notification-destination/", "/api/v1/external-role", "/api/v1/meta/", "/api/v1/debug/", "/api/v1/datasource/")
    if not path.startswith(global_routes):
        parameters.append({"name": "X-Namespace", "in": "header", "required": True, "schema": {"type": "string", "format": "uuid"}})
    operation_data = {"operationId": operation, "parameters": parameters,
                      "responses": {"200": {"description": "Success"}}}
    if output != "void":
        operation_data["responses"]["200"]["content"] = {"application/json": {"schema": schema(output)}}
    payload = re.search(r"JSON.stringify\((\w+)\)", body)
    if payload and payload.group(1) in argument_types:
        operation_data["requestBody"] = {"required": True, "content": {"application/json": {"schema": schema(argument_types[payload.group(1)])}}}
    if path == "/api/v1/user/login":
        operation_data["requestBody"] = {"required": True, "content": {"application/json": {"schema": {
            "type": "object", "required": ["email", "password"],
            "properties": {"email": {"type": "string"}, "password": {"type": "string", "format": "password"}}}}}}
    if path in ("/api/v1/user/login", "/api/v1/user/login-code-flow", "/api/v1/user/callback"):
        operation_data["security"] = []
    for status in ("400", "401", "403", "404", "409", "429", "500", "502", "504"):
        operation_data["responses"][status] = {"description": "Request failed"}
    if path in ("/api/v1/user/login-code-flow", "/api/v1/user/callback"):
        operation_data["responses"].pop("200", None)
        operation_data["responses"]["302"] = {"description": "Continue login in the browser"}
    if path == "/api/v1/meta/database-schema":
        operation_data["responses"]["200"]["content"] = {"application/json": {"schema": {"type": "array", "items": {"type": "object", "properties": {"schema": {"type": "string"}, "tables": {"type": "array", "items": {"type": "string"}}}}}}}
    if path.startswith("/api/v1/debug/rollback/"):
        operation_data["responses"].pop("200", None)
    paths.setdefault(path, {})[verb] = operation_data

document = {"openapi": "3.0.3", "info": {"title": "DataCat API", "version": "1.0"},
            "security": [{"bearerAuth": []}, {"cookieAuth": []}], "paths": paths,
            "components": {"securitySchemes": {"bearerAuth": {"type": "http", "scheme": "bearer"},
                                                 "cookieAuth": {"type": "apiKey", "in": "cookie", "name": "access_token"}},
                           "schemas": schemas}}
(ROOT / "api/openapi.json").write_text(json.dumps(document, indent=2) + "\n")
print(f"Exported {sum(len(methods) for methods in paths.values())} operations and {len(schemas)} schemas")
