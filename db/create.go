// Copyright © 2016 Abcum Ltd
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package db

import (
	"github.com/abcum/surreal/kvs"
	"github.com/abcum/surreal/sql"
	"github.com/abcum/surreal/util/item"
	"github.com/abcum/surreal/util/keys"
	"github.com/abcum/surreal/util/uuid"
)

func executeCreateStatement(ast *sql.CreateStatement) (out []interface{}, err error) {

	txn, err := db.Txn(true)
	if err != nil {
		return
	}

	defer txn.Rollback()

	for _, w := range ast.What {

		if what, ok := w.(*sql.Thing); ok {
			key := &keys.Thing{KV: ast.KV, NS: ast.NS, DB: ast.DB, TB: what.TB, ID: what.ID}
			kv, _ := txn.Get(key.Encode())
			doc := item.New(kv, key)
			if ret, err := create(txn, doc, ast); err != nil {
				return nil, err
			} else if ret != nil {
				out = append(out, ret)
			}
		}

		if what, ok := w.(sql.Table); ok {
			key := &keys.Thing{KV: ast.KV, NS: ast.NS, DB: ast.DB, TB: what, ID: uuid.NewV5(uuid.NewV4().UUID, ast.KV).String()}
			kv, _ := txn.Get(key.Encode())
			doc := item.New(kv, key)
			if ret, err := create(txn, doc, ast); err != nil {
				return nil, err
			} else if ret != nil {
				out = append(out, ret)
			}
		}

	}

	txn.Commit()

	return

}

func create(txn *kvs.TX, doc *item.Doc, ast *sql.CreateStatement) (out interface{}, err error) {

	if err = doc.Merge(txn, ast.Data); err != nil {
		return nil, err
	}

	if err = doc.StoreIndex(txn); err != nil {
		return nil, err
	}

	if err = doc.StartThing(txn); err != nil {
		return nil, err
	}

	if err = doc.StorePatch(txn); err != nil {
		return nil, err
	}

	if err = doc.StoreTrail(txn); err != nil {
		return nil, err
	}

	out = doc.Yield(ast.Echo, sql.AFTER)

	return

}
