/*
 * Copyright (C) 2019 The themis Authors
 * This file is part of The themis library.
 *
 * The themis is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The themis is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The themis.  If not, see <http://www.gnu.org/licenses/>.
 */

// Package error privides error code for http
package error

import ontErrors "github.com/saveio/themis/errors"

const (
	SUCCESS            int64 = 0
	SESSION_EXPIRED    int64 = 41001
	SERVICE_CEILING    int64 = 41002
	ILLEGAL_DATAFORMAT int64 = 41003
	INVALID_VERSION    int64 = 41004

	INVALID_METHOD int64 = 42001
	INVALID_PARAMS int64 = 42002

	INVALID_TRANSACTION int64 = 43001
	INVALID_ASSET       int64 = 43002
	INVALID_BLOCK       int64 = 43003

	UNKNOWN_TRANSACTION int64 = 44001
	UNKNOWN_ASSET       int64 = 44002
	UNKNOWN_BLOCK       int64 = 44003
	UNKNOWN_CONTRACT    int64 = 44004

	INTERNAL_ERROR       int64 = 45001
	SMARTCODE_ERROR      int64 = 47001
	PRE_EXEC_ERROR       int64 = 47002
	JSON_MARSHAL_ERROR   int64 = 47003
	JSON_UNMARSHAL_ERROR int64 = 47004
	SIG_VERIFY_ERROR     int64 = 47005

	ENDPOINT_NOT_FOUND int64 = 48001

	CHANNEL_TARGET_HOST_INFO_NOT_FOUND int64 = 49001
	BLOCK_SYNCING_UNCOMPLETE           int64 = 49002

	CHANNEL_ERROR int64  = 50000
	PASSWORD_WRONG int64 = 50001
)

var ErrMap = map[int64]string{
	SUCCESS:            "SUCCESS",
	SESSION_EXPIRED:    "SESSION EXPIRED",
	SERVICE_CEILING:    "SERVICE CEILING",
	ILLEGAL_DATAFORMAT: "ILLEGAL DATAFORMAT",
	INVALID_VERSION:    "INVALID VERSION",

	INVALID_METHOD: "INVALID METHOD",
	INVALID_PARAMS: "INVALID PARAMS",

	INVALID_TRANSACTION: "INVALID TRANSACTION",
	INVALID_ASSET:       "INVALID ASSET",
	INVALID_BLOCK:       "INVALID BLOCK",

	UNKNOWN_TRANSACTION: "UNKNOWN TRANSACTION",
	UNKNOWN_ASSET:       "UNKNOWN ASSET",
	UNKNOWN_BLOCK:       "UNKNOWN BLOCK",
	UNKNOWN_CONTRACT:    "UNKNOWN CONTRACT",

	INTERNAL_ERROR:       "INTERNAL ERROR",
	SMARTCODE_ERROR:      "SMARTCODE EXEC ERROR",
	PRE_EXEC_ERROR:       "SMARTCODE PREPARE EXEC ERROR",
	JSON_MARSHAL_ERROR:   "json.Marshal ERROR",
	JSON_UNMARSHAL_ERROR: "json.Unmarshal ERROR",

	ENDPOINT_NOT_FOUND: "ENDPOINT NOT FOUND",

	CHANNEL_TARGET_HOST_INFO_NOT_FOUND: "CHANNEL TARGET HOST INFO NOT FOUND",
	BLOCK_SYNCING_UNCOMPLETE:           "BLOCK SYNCING UNCOMPLETE",

	CHANNEL_ERROR:  "CHANNEL OPERATING ERROR",
	PASSWORD_WRONG: "INCORRECT PASSWORD",

	int64(ontErrors.ErrNoCode):               "INTERNAL ERROR, ErrNoCode",
	int64(ontErrors.ErrUnknown):              "INTERNAL ERROR, ErrUnknown",
	int64(ontErrors.ErrDuplicatedTx):         "INTERNAL ERROR, ErrDuplicatedTx",
	int64(ontErrors.ErrDuplicateInput):       "INTERNAL ERROR, ErrDuplicateInput",
	int64(ontErrors.ErrAssetPrecision):       "INTERNAL ERROR, ErrAssetPrecision",
	int64(ontErrors.ErrTransactionBalance):   "INTERNAL ERROR, ErrTransactionBalance",
	int64(ontErrors.ErrAttributeProgram):     "INTERNAL ERROR, ErrAttributeProgram",
	int64(ontErrors.ErrTransactionContracts): "INTERNAL ERROR, ErrTransactionContracts",
	int64(ontErrors.ErrTransactionPayload):   "INTERNAL ERROR, ErrTransactionPayload",
	int64(ontErrors.ErrDoubleSpend):          "INTERNAL ERROR, ErrDoubleSpend",
	int64(ontErrors.ErrTxHashDuplicate):      "INTERNAL ERROR, ErrTxHashDuplicate",
	int64(ontErrors.ErrStateUpdaterVaild):    "INTERNAL ERROR, ErrStateUpdaterVaild",
	int64(ontErrors.ErrSummaryAsset):         "INTERNAL ERROR, ErrSummaryAsset",
	int64(ontErrors.ErrXmitFail):             "INTERNAL ERROR, ErrXmitFail",
	int64(ontErrors.ErrNoAccount):            "INTERNAL ERROR, ErrNoAccount",
}
