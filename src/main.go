package main

// YouTube audio player service.
// This service deals with,
//   * searching
//   * conversion from video to audio
//   * successive distribution
func main() {
}

// GET /
// index.html <- single page driven by React

// ====================================================================================================
// Resource: videos
// Desc: video information such as title, thumbnail, etc

// GET /videos?q=hoge
// {
// 	"hit_count": "30",
// 	"videos": [
// 		{
// 			"id": "a30jvlkjs03",
// 			"title": "hoge",
// 			"duration": "PT1H43M31S",
// 			"view_count": "136421",
// 			"publish_data": "2018-03-29",
// 			"thumbnail": {
// 				"default": {
// 					"url": "https//..." ,
// 					"width": "320",
// 					"height": "180",
// 				}
// 				"medium": {
// 					"url": "https//..." ,
// 					"width": "480",
// 					"height": "360",
// 				}
// 				"standard": {
// 					"url": "https//..." ,
// 					"width": "640",
// 					"height": "480",
// 				}
// 			}
// 		},...
// 	]
// }

// GET /videos?relatedID=a30jvlkjs03
// {
// 	"hit_count": "30",
// 	"videos": [
// 		{
// 			"id": "alskjeo93-s",
// 			"title": "hoge",
// 			"duration": "PT1H43M31S",
// 			"view_count": "136421",
// 			"publish_data": "2018-03-29",
// 			"thumbnail": {
// 				"default": {
// 					"url": "https//..." ,
// 					"width": "320",
// 					"height": "180",
// 				}
// 				"medium": {
// 					"url": "https//..." ,
// 					"width": "480",
// 					"height": "360",
// 				}
// 				"standard": {
// 					"url": "https//..." ,
// 					"width": "640",
// 					"height": "480",
// 				}
// 			}
// 		},...
// 	]
// }

// GET /videos/:id
// {
// 	"id": "a30jvlkjs03",
// 	"title": "hoge",
// 	"duration": "PT1H43M31S",
// 	"view_count": "136421",
// 	"publish_data": "2018-03-29",
// 	"thumbnail": {
// 		"default": {
// 			"url": "https//..." ,
// 			"width": "320",
// 			"height": "180",
// 		}
// 		"medium": {
// 			"url": "https//..." ,
// 			"width": "480",
// 			"height": "360",
// 		}
// 		"standard": {
// 			"url": "https//..." ,
// 			"width": "640",
// 			"height": "480",
// 		}
// 	}
// }
// ====================================================================================================

// ====================================================================================================
// Resource: streams
// Desc: Returns a url of HLS segment list file

// GET /streams/:id
// {
// 	"id": "a30jvlkjs03",
// 	"segment_list_file_url": "/.../a30jvlkjs03/audio.m3u8",  <- relative path from index url
// }
// ====================================================================================================

// ====================================================================================================
// Resource: users
// Desc: Manage user information (session management)

// GET /users/:id <- authentication is required
// {
//	"user": {
//		"name": "Matthew Ishige",
//		"email": "matthewlujp@gmail.com",
//		"uid": "laksa09lksj",
//	},
// 	"recently_played": {
// 		"count": "30",
// 		"videos": [
// 			{
// 				"id": "a30jvlkjs03",
// 				"title": "hoge",
// 				"duration": "PT1H43M31S",
// 				"view_count": "136421",
// 				"publish_data": "2018-03-29",
// 				"thumbnail": {
// 					"default": {
// 						"url": "https//..." ,
// 						"width": "320",
// 						"height": "180",
// 					}
// 					"medium": {
// 						"url": "https//..." ,
// 						"width": "480",
// 						"height": "360",
// 					}
// 					"standard": {
// 						"url": "https//..." ,
// 						"width": "640",
// 						"height": "480",
// 					}
// 				}
// 			},...
// 		]
// 	}
// }

// POST /users
// DATA
// {
// 	"name": "Matthew Ishige",
// 	"email": "matthewlujp@gmail.com",
// 	"passwd": "abcdefg",
// }
// RESPONSE
// {
// 	"message": "Welcome!",
//	"user": {
//		"name": "Matthew Ishige",
//		"email": "matthewlujp@gmail.com",
//		"uid": "laksa09lksj",
//	}
// }

// UPDATE /users
// DATA
// {
// 	"name": "Matthew Lu",
// }
// RESPONSE
// {
// 	"message": "Update succeeded."
//	"user": {
//		"name": "Matthew Ishige",
//		"email": "matthewlujp@gmail.com",
//		"uid": "laksa09lksj",
//	}
// }
// ====================================================================================================

// ====================================================================================================
// Resource: login
// Desc: Manage loging process. Session expire mechanism is necessary.

// POST /login
// DATA
// {
// 	"email": "matthewlujp@gmail.com",
// 	"passwd": "foobar",
// }
// RESPONSE
// {
// 	"message": "login succeeded",
// 	"user": {
// 		"name": "Matthew Ishige",
// 		"email": "matthewlujp@gmail.com",
// 		"uid": "laksa09lksj",
// 	}
// }
