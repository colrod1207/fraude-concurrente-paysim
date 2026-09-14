"""Tipos y constantes del dataset PaySim."""

from dataclasses import dataclass
from typing import Dict, List, Optional

EXPECTED_COLUMNS = 11


class DiscardReason:
    EMPTY_ROW = "empty_row"
    WRONG_COLUMN_COUNT = "wrong_column_count"
    PARSE_ERROR = "parse_error"
    INVALID_TYPE = "invalid_type"
    NEGATIVE_AMOUNT = "negative_amount"
    IMPOSSIBLE_BALANCE = "impossible_balance"


VALID_TYPES: Dict[str, int] = {
    "CASH_IN": 0,
    "CASH_OUT": 1,
    "DEBIT": 2,
    "PAYMENT": 3,
    "TRANSFER": 4,
}


@dataclass
class CleanRecord:
    step: int
    type: str
    type_code: int
    amount: float
    name_orig: str
    old_balance_org: float
    new_balance_orig: float
    name_dest: str
    old_balance_dest: float
    new_balance_dest: float
    is_fraud: int
    is_flagged_fraud: int
    error_balance_orig: float
    error_balance_dest: float
    is_merchant_dest: int
    hour: int


@dataclass
class Result:
    record: Optional[CleanRecord] = None
    discard: Optional[str] = None


def header() -> List[str]:
    return [
        "step", "type", "typeCode", "amount",
        "nameOrig", "oldbalanceOrg", "newbalanceOrig",
        "nameDest", "oldbalanceDest", "newbalanceDest",
        "isFraud", "isFlaggedFraud",
        "errorBalanceOrig", "errorBalanceDest",
        "isMerchantDest", "hour",
    ]
