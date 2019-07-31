import http from "http";
import crypto from "crypto";

import shell from "shelljs";
import R from "ramda";

const PORT: number = Number(process.env.port) || 5555; // server port
const KEY: string = process.env.secret || ""; // write your secret key here
const BLOG: string = process.env.place || ""; // blog place

/**
 * judge if two buffer string same
 * @param a buffer string
 * @param b buffer string
 */
function eq(a: Buffer, b: Buffer): boolean {
  if (a.length !== b.length) {
    return false;
  }

  let res = 0;
  for (let i: number = 0; i < a.length; ++i) {
    res |= a[i] ^ b[i];
  }
  return res === 0;
}

/**
 * check data in buffer
 * @param sign get the data from request header
 * @param data get the response data to check the header
 */
function bufferEqual(sign: any, data: any): boolean {
  return eq(
    Buffer.from(sign),
    Buffer.from(
      "sha1=".concat(
        crypto
          .createHmac("sha1", KEY)
          .update(data)
          .digest("hex")
      )
    )
  );
}

// create server
http
  .createServer(function(req, res) {
    req.on("data", async function(data: Buffer) {
      try {
        if (bufferEqual(req.headers["x-hub-signature"], data)) {
          let gitName = JSON.parse(data.toString())["repository"]["clone_url"];
          let repoName = JSON.parse(data.toString())["repository"]["name"];
          shell.cd("/tmp");
          shell.exec("git clone " + gitName);
          shell.mv("/tmp/" + repoName + "/md/*", BLOG + "/content/post"); // move the md posts into content post
          shell.mv("/tmp/" + repoName + "/img/*", BLOG + "/static");

          shell.exec("systemctl stop nginx.service");
          shell.cd(BLOG);
          shell.rm("-rf", "public");
          shell.exec("hugo");
          shell.exec("systemctl start nginx.service");
        } else {
          // FIXME: need to log in the server
        }
      } catch (error) {
        console.log(error);
      }
    });
    res.end();
  })
  .listen(PORT);
