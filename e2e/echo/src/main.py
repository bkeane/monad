import uvicorn
from boto3 import client
from os import getenv, environ
from fastapi import FastAPI, Request, HTTPException
from pydantic import ValidationError
from fastapi.responses import PlainTextResponse
from models import GetFunctionResponse, GetRoleResponse, ListAttachedRolePoliciesResponse, AttachedPolicy, GetPolicyResponse, EventBridgeEvent
from cloudwatch import get_latest_logs
import logging

# Configure logging
logger = logging.getLogger("uvicorn")

app = FastAPI(
    title="Echo",
    description="Echo is an introspection service for assisting e2e tests.",
    version="0.0.1",
    docs_url="/public/docs",
    openapi_url="/public/openapi.json",
    debug=True
)

@app.middleware("http")
async def set_root_path(request: Request, call_next):
    if request.headers.get("x-forwarded-prefix"):
        app.root_path = request.headers.get("x-forwarded-prefix")
    response = await call_next(request)
    return response

@app.middleware("http")
async def log_incoming_events(request: Request, call_next):
    if request.url.path == "/events":
        body = await request.body()
        logger.info("EVENT: %s", body.decode('utf-8'))
    
    response = await call_next(request)
    return response

@app.middleware("http")
async def configure_boto3(request: Request, call_next):
    request.state.function_name = getenv("AWS_LAMBDA_FUNCTION_NAME")
    request.state.lambdac = client('lambda')
    request.state.iamc = client('iam')
    request.state.logc = client('logs')
    return await call_next(request)

@app.get("/health")
async def health():
    return {"status": "ok"}

@app.get("/public/health")
async def public_health():
    return {"status": "ok"}

@app.post("/events")
async def events(event: EventBridgeEvent):
    logger.info("Received event: %s", event.model_dump_json())
    return {"status": "ok"}

@app.get("/headers")
async def headers(request: Request):
    return request.headers

@app.get("/headers/{key}", response_class=PlainTextResponse)
async def header(key: str, request: Request):
    if key not in request.headers.keys():
        raise HTTPException(status_code=404, detail="header not found")
    
    return request.headers.get(key)

@app.get('/env')
async def env():
    return environ

@app.get("/env/{key}", response_class=PlainTextResponse)
async def env(key: str):
    if key not in environ.keys():
        raise HTTPException(status_code=404, detail="environment variable not found")
    
    return environ[key]

@app.get("/function")
async def get_function(request: Request) -> GetFunctionResponse:
    response = request.state.lambdac.get_function(
        FunctionName=request.state.function_name
    )

    try:
        validated = GetFunctionResponse.model_validate(response, strict=False)
    except ValidationError as e:
        raise HTTPException(status_code=422, detail=f"Validation error: {str(e)}")
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Unexpected error: {str(e)}")  
    
    return validated

@app.get("/function/name", response_class=PlainTextResponse)
async def get_function_name(request: Request) -> str:
    validated = await get_function(request)
    return validated.Configuration.FunctionName

@app.get("/function/role")
async def get_function_role(request: Request) -> GetRoleResponse:
    function = await get_function(request)

    if function.Configuration.Role is None:
        raise HTTPException(status_code=404, detail="role not found")
    
    role_parts = function.Configuration.Role.split("role/")
    if len(role_parts) != 2:
        raise HTTPException(status_code=422, detail="invalid role ARN format")

    response = request.state.iamc.get_role(
        RoleName=role_parts[1]
    )

    try:
        validated = GetRoleResponse.model_validate(response, strict=False)
    except ValidationError as e:
        raise HTTPException(status_code=422, detail=f"Validation error: {str(e)}")
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Unexpected error: {str(e)}")
    
    return validated

@app.get("/function/role/name", response_class=PlainTextResponse)
async def get_function_role_name(request: Request) -> str:
    role = await get_function_role(request)
    return role.Role.RoleName

@app.get("/function/role/tags")
async def get_function_role_tags(request: Request) -> dict[str, str]:
    role = await get_function_role(request)
    tag_dict = {tag.Key: tag.Value for tag in role.Role.Tags}
    return tag_dict

@app.get("/function/role/tags/{key}", response_class=PlainTextResponse)
async def get_function_role_tag(request: Request, key: str) -> str:
    tags = await get_function_role_tags(request)

    if key not in tags.keys():
        raise HTTPException(status_code=404, detail="tag not found")
    
    return tags[key]

@app.get("/function/role/policies")
async def get_function_role_policies(request: Request) -> list[AttachedPolicy]:
    role = await get_function_role(request)
    response = request.state.iamc.list_attached_role_policies(
        RoleName=role.Role.RoleName
    )
    
    try:
        validated = ListAttachedRolePoliciesResponse.model_validate(response, strict=False)
    except ValidationError as e:
        raise HTTPException(status_code=422, detail=f"Validation error: {str(e)}")
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Unexpected error: {str(e)}")
    
    return validated.AttachedPolicies

@app.get("/function/role/policies/{policy_name}")
async def get_function_role_policy(request: Request, policy_name: str) -> GetPolicyResponse:
    policies = await get_function_role_policies(request)

    for policy in policies:
        if policy.PolicyName == policy_name:
            response = request.state.iamc.get_policy(
                PolicyArn=policy.PolicyArn
            )

            try:
                validated = GetPolicyResponse.model_validate(response, strict=False)
            except ValidationError as e:
                raise HTTPException(status_code=422, detail=f"Validation error: {str(e)}")
            except Exception as e:
                raise HTTPException(status_code=500, detail=f"Unexpected error: {str(e)}")
            
            return validated

    raise HTTPException(status_code=404, detail="policy not found")

@app.get("/function/role/policies/{policy_name}/tags")
async def get_function_role_policy_tags(request: Request, policy_name: str) -> dict[str, str]:
    policy = await get_function_role_policy(request, policy_name)
    tag_dict = {tag.Key: tag.Value for tag in policy.Policy.Tags}
    return tag_dict

@app.get("/function/tags")
async def get_function_tags(request: Request) -> dict[str, str]:
    validated = await get_function(request)
    return validated.Tags

@app.get("/function/tags/{key}", response_class=PlainTextResponse)
async def get_function_tag(request: Request, key: str) -> str:
    validated = await get_function_tags(request)

    if key not in validated.keys():
        raise HTTPException(status_code=404, detail="tag not found")
    
    return validated[key]

@app.get("/function/memory", response_class=PlainTextResponse)
async def get_function_configuration_memory(request: Request) -> str:
    validated = await get_function(request)
    return str(validated.Configuration.MemorySize)

@app.get("/function/timeout", response_class=PlainTextResponse)
async def get_function_configuration_timeout(request: Request) -> str:
    validated = await get_function(request)
    return str(validated.Configuration.Timeout)

@app.get("/function/disk", response_class=PlainTextResponse)
async def get_function_configuration_disk(request: Request) -> str:
    validated = await get_function(request)
    return str(validated.Configuration.EphemeralStorage.Size)

@app.get("/function/log_group", response_class=PlainTextResponse)
async def get_function_configuration_log_group(request: Request) -> str:
    validated = await get_function(request)
    return validated.Configuration.LoggingConfig.LogGroup

@app.get("/function/log_group/tail")
async def tail_function_log_group(request: Request, n: int = 10, grep: str = None, expect: bool = False) -> list[str]:
    log_group = await get_function_configuration_log_group(request)
    logs = get_latest_logs(request.state.logc, log_group, n, request.url.path, grep)

    # If expect is true, return a 404 if no logs are found
    if expect:
        if len(logs) == 0:
            raise HTTPException(status_code=404, detail="log not found")
        
    return [log.message for log in logs]

if __name__ == '__main__':
    uvicorn.run(app, host="0.0.0.0", port=int(getenv("PORT","8090")))
