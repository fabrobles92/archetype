// SPDX-License-Identifier: MIT
pragma solidity >=0.8.0;

/* Autogenerated file. Do not edit manually. */

enum TableId {
    Tick,
    Move
}

struct RowData_Move {
    uint8 playerId;
    uint8 direction;
}

interface ITableGetter {
    function getTickRow(
    ) external view returns (RowData_Tick memory);

    function getMoveRow(
    ) external view returns (RowData_Move memory);

}
