#include <sqlite3.h>
#include <SPIFFS.h>
#include <map>
#include <string>
#include <vector>
#include <optional>

#define FORMAT_SPIFFS_IF_FAILED true

namespace repo {

    static sqlite3 *db = nullptr;
    static int err = 0;

    void open(std::string filename) {
        if (!SPIFFS.begin(FORMAT_SPIFFS_IF_FAILED)) {
            Serial.println("Failed to mount file system");
            return;
        }
        SPIFFS.remove("/test1.db");

        sqlite3_initialize();
        err = sqlite3_open(filename.c_str(), &db);
        if (err) {
            Serial.printf("Can't open database: %s\n", sqlite3_errmsg(db));
            return;
        }
        Serial.println("Opened database successfully");

        std::string sql = "CREATE TABLE IF NOT EXISTS post (id INTEGER PRIMARY KEY, title TEXT, content TEXT, parent_id INTEGER);";
        char *errmsg;
        err = sqlite3_exec(db, sql.c_str(), nullptr, nullptr, &errmsg);
        if (err != SQLITE_OK) {
            Serial.printf("SQL error: %s\n", sqlite3_errmsg(db));
            Serial.printf("error message: %s\n", errmsg);
            Serial.printf("error code: %d\n", err);
            return;
        }
    }

    typedef int Row_getReplyIds;

    std::vector<Row_getReplyIds> getReplyIds(int parent_id){
        std::string sql = "SELECT id FROM post WHERE parent_id = ?;";
        sqlite3_stmt* stmt = nullptr;
        err = sqlite3_prepare_v2(db, sql.c_str(), -1, &stmt, nullptr);
        if (err != SQLITE_OK) {
            Serial.printf("SQL error: %s\n", sqlite3_errmsg(db));
            return; // fix return
        }
        err = sqlite3_bind_int(stmt, 1, parent_id);
        if (err != SQLITE_OK) {
            Serial.printf("SQL error: %s\n", sqlite3_errmsg(db));
            return;
        }
        std::vector<Row_getReplyIds> result;
        while (sqlite3_step(stmt) == SQLITE_ROW) {
            Row_getReplyIds row;
            row = sqlite3_column_int(stmt, 0);
            result.push_back(row);
        }
        sqlite3_finalize(stmt);
        return result;
    }

    typedef struct {
        int id;
        std::string title;
        std::string content;
        std::optional<int> parent_id;
    } Row_getAllPosts;

    std::vector<Row_getAllPosts> getAllPosts(){
        std::string sql = "SELECT id, title, content, parent_id FROM post;";
        sqlite3_stmt* stmt = nullptr;
        err = sqlite3_prepare_v2(db, sql.c_str(), -1, &stmt, nullptr);
        if (err != SQLITE_OK) {
            Serial.printf("SQL error: %s\n", sqlite3_errmsg(db));
            return {};
        }
        std::vector<Row_getAllPosts> result;
        while (sqlite3_step(stmt) == SQLITE_ROW) {
            Row_getAllPosts row;
            row.id = sqlite3_column_int(stmt, 0);
            row.title = std::string(reinterpret_cast<const char*>(sqlite3_column_text(stmt, 1)));
            row.content = std::string(reinterpret_cast<const char*>(sqlite3_column_text(stmt, 2)));
            if (sqlite3_column_type(stmt, 3) != SQLITE_NULL) {
                row.parent_id.emplace(sqlite3_column_int(stmt, 3)); // checar se é válido
            }
            result.push_back(row);
        }
        sqlite3_finalize(stmt);
        return result;
    }

    void updatePost(
        std::string title,
        std::string content,
        int id
    ){
        std::string sql = "UPDATE post SET title = ?, content = ? WHERE id = ?;";
        sqlite3_stmt* stmt = nullptr;
        err = sqlite3_prepare_v2(db, sql.c_str(), -1, &stmt, nullptr);
        if (err != SQLITE_OK) {
            Serial.printf("SQL error: %s\n", sqlite3_errmsg(db));
            return;
        }
        err = sqlite3_bind_text(stmt, 1, title.c_str(), -1, SQLITE_STATIC);
        if (err != SQLITE_OK) {
            Serial.printf("SQL error: %s\n", sqlite3_errmsg(db));
            return;
        }
        err = sqlite3_bind_text(stmt, 2, content.c_str(), -1, SQLITE_STATIC);
        if (err != SQLITE_OK) {
            Serial.printf("SQL error: %s\n", sqlite3_errmsg(db));
            return;
        }
        err = sqlite3_bind_int(stmt, 3, id);
        if (err != SQLITE_OK) {
            Serial.printf("SQL error: %s\n", sqlite3_errmsg(db));
            return;
        }
        err = sqlite3_step(stmt);
        if (err != SQLITE_DONE) {
            Serial.printf("SQL error: %s\n", sqlite3_errmsg(db));
            return;
        }
        sqlite3_finalize(stmt);
    }

    void createPost(
        std::string title,
        std::string content,
        std::optional<int> parent_id
    ){
        std::string sql = "INSERT INTO post (title, content, parent_id) VALUES (?, ?, ?);";
        sqlite3_stmt* stmt = nullptr;
        err = sqlite3_prepare_v2(db, sql.c_str(), -1, &stmt, nullptr);
        if (err != SQLITE_OK) {
            Serial.printf("SQL error: %s\n", sqlite3_errmsg(db));
            return;
        }
        sqlite3_bind_text(stmt, 1, title.c_str(), -1, SQLITE_STATIC);
        if (err != SQLITE_OK) {
            Serial.printf("SQL error: %s\n", sqlite3_errmsg(db));
            return;
        }
        sqlite3_bind_text(stmt, 2, content.c_str(), -1, SQLITE_STATIC);
        if (err != SQLITE_OK) {
            Serial.printf("SQL error: %s\n", sqlite3_errmsg(db));
            return;
        }
        if (parent_id.has_value()) {
            sqlite3_bind_int(stmt, 3, parent_id.value());
        } else {
            sqlite3_bind_null(stmt, 3);
        }
        err = sqlite3_step(stmt);
        if (err != SQLITE_DONE) {
            Serial.printf("SQL error: %s\n", sqlite3_errmsg(db));
            return;
        }
        sqlite3_finalize(stmt);
    }

}
