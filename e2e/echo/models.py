from pydantic import BaseModel
from enum import Enum
class EphemeralStorage(BaseModel):
    Size: int

class VpcConfig(BaseModel):
    SecurityGroups: list[str]
    Subnets: list[str]

class Code(BaseModel):
    ImageUri: str

class Architecture(str, Enum):
    x86_64 = "x86_64"
    arm64 = "arm64"

class LoggingConfig(BaseModel):
    LogFormat: str
    LogGroup: str

class Configuration(BaseModel):
    FunctionName: str
    Timeout: int
    MemorySize: int
    EphemeralStorage: EphemeralStorage
    Role: str
    LoggingConfig: LoggingConfig
    
class GetFunctionResponse(BaseModel):
    Configuration: Configuration
    Tags: dict[str, str]

class Tag(BaseModel):
    Key: str
    Value: str

class Role(BaseModel):
    RoleName: str
    Tags: list[Tag]

class GetRoleResponse(BaseModel):
    Role: Role

class AttachedPolicy(BaseModel):
    PolicyName: str
    PolicyArn: str

class ListAttachedRolePoliciesResponse(BaseModel):
    AttachedPolicies: list[AttachedPolicy]

class Policy(BaseModel):
    PolicyName: str
    Arn: str
    Tags: list[Tag]

class GetPolicyResponse(BaseModel):
    Policy: Policy